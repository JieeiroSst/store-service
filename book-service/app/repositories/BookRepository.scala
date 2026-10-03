package repositories

import java.sql.Connection
import java.time.Instant
import javax.inject._

import anorm.SqlParser._
import anorm._
import models._
import play.api.db.Database

import scala.concurrent.Future

/**
 * Keyset (cursor) pagination: each page is fetched with
 * `WHERE key > / < cursor ORDER BY key LIMIT n + 1`, so the cost does not
 * grow with the page depth (unlike OFFSET) and rows inserted meanwhile
 * do not shift or duplicate results.
 */
@Singleton
class BookRepository @Inject() (db: Database, categories: CategoryRepository)(implicit ec: DatabaseExecutionContext) {

  private val BookColumns =
    "b.id, b.title, b.author, b.isbn, b.description, b.published_year, b.cover_key, b.cover_content_type, " +
      "b.created_at, b.updated_at"

  private val bookParser: RowParser[BookRow] =
    (long("id") ~ str("title") ~ str("author") ~ get[Option[String]]("isbn") ~
      get[Option[String]]("description") ~ get[Option[Int]]("published_year") ~
      get[Option[String]]("cover_key") ~ get[Option[String]]("cover_content_type") ~
      get[Instant]("created_at") ~ get[Instant]("updated_at")).map {
      case id ~ title ~ author ~ isbn ~ description ~ year ~ coverKey ~ coverType ~ createdAt ~ updatedAt =>
        BookRow(id, title, author, isbn, description, year, coverKey, coverType, createdAt, updatedAt)
    }

  private val ChapterColumns = "book_id, chapter_number, title, content_key, content_type, size_bytes"

  private val chapterParser: RowParser[ChapterRef] =
    (long("book_id") ~ int("chapter_number") ~ str("title") ~ str("content_key") ~ str("content_type") ~
      get[Option[Long]]("size_bytes")).map { case bookId ~ number ~ title ~ key ~ contentType ~ size =>
      ChapterRef(bookId, number, title, key, contentType, size)
    }

  // ---- Books ---------------------------------------------------------------

  /**
   * Books matching `filter`, newest first. The cursor is the id of the last
   * book returned, so filters and the cursor combine into one index range.
   */
  def list(filter: BookFilter, page: PageRequest): Future[Page[BookRow]] = Future {
    val conditions = List.newBuilder[String]
    val params     = List.newBuilder[NamedParameter]

    filter.q.foreach { q =>
      conditions += "(LOWER(b.title) LIKE {q} OR LOWER(b.author) LIKE {q})"
      params += ("q" -> SqlUtil.likeContains(q))
    }
    filter.author.foreach { a =>
      conditions += "LOWER(b.author) = {author}"
      params += ("author" -> a.toLowerCase)
    }
    filter.year.foreach { y =>
      conditions += "b.published_year = {year}"
      params += ("year" -> y)
    }
    filter.category.foreach { slug =>
      conditions += """EXISTS (SELECT 1 FROM book_categories bc JOIN categories c ON c.id = bc.category_id
                      |        WHERE bc.book_id = b.id AND c.slug = {category})""".stripMargin
      params += ("category" -> slug)
    }
    page.after.foreach { afterId =>
      conditions += "b.id < {after}"
      params += ("after" -> afterId)
    }
    params += ("fetch" -> (page.limit + 1))

    val where = conditions.result() match {
      case Nil => ""
      case cs  => cs.mkString("WHERE ", " AND ", "")
    }
    val rows = db.withConnection { implicit c =>
      SQL(s"SELECT $BookColumns FROM books b $where ORDER BY b.id DESC LIMIT {fetch}")
        .on(params.result(): _*)
        .as(bookParser.*)
    }
    Page.fromFetched(rows, page.limit)(_.id)
  }

  /** The book and its number of chapters. */
  def find(id: Long): Future[Option[(BookRow, Int)]] = Future {
    db.withConnection { implicit c =>
      findRow(id).map { book =>
        book -> SQL"SELECT COUNT(*) FROM book_chapters WHERE book_id = $id".as(scalar[Int].single)
      }
    }
  }

  def findByIsbn(isbn: String): Future[Option[BookRow]] = Future {
    db.withConnection { implicit c =>
      SQL(s"SELECT $BookColumns FROM books b WHERE b.isbn = {isbn}").on("isbn" -> isbn).as(bookParser.singleOpt)
    }
  }

  /** Categories of each book, for building a page of results in one query. */
  def categoriesFor(bookIds: Seq[Long]): Future[Map[Long, List[Category]]] =
    if (bookIds.isEmpty) Future.successful(Map.empty)
    else
      Future {
        db.withConnection { implicit c =>
          SQL("""SELECT bc.book_id, c.id, c.slug, c.name FROM book_categories bc
                |JOIN categories c ON c.id = bc.category_id
                |WHERE bc.book_id IN ({ids}) ORDER BY c.name""".stripMargin)
            .on("ids" -> bookIds)
            .as((long("book_id") ~ categories.parser).*)
            .groupMap(_._1)(_._2)
        }
      }

  /** Left(message) when the ISBN belongs to another book. */
  def insert(input: BookInput, categoryIds: Seq[Long]): Future[Either[String, Long]] = Future {
    uniqueIsbn(input) {
      db.withTransaction { implicit c =>
        val id = SQL"""INSERT INTO books (title, author, isbn, description, published_year)
                       VALUES (${input.title}, ${input.author}, ${input.isbn}, ${input.description},
                               ${input.publishedYear})""".executeInsert(get[Long](1).single)
        linkCategories(id, categoryIds)
        id
      }
    }
  }

  /** Replaces the book's fields and categories. Right(false) when it does not exist. */
  def update(id: Long, input: BookInput, categoryIds: Seq[Long]): Future[Either[String, Boolean]] = Future {
    uniqueIsbn(input) {
      db.withTransaction { implicit c =>
        val updated = SQL"""UPDATE books SET title = ${input.title}, author = ${input.author}, isbn = ${input.isbn},
                              description = ${input.description}, published_year = ${input.publishedYear},
                              updated_at = CURRENT_TIMESTAMP
                            WHERE id = $id""".executeUpdate() > 0
        if (updated) {
          SQL"DELETE FROM book_categories WHERE book_id = $id".executeUpdate()
          linkCategories(id, categoryIds)
        }
        updated
      }
    }
  }

  /**
   * Deletes the book with its chapters and category links. Returns the object
   * keys it referenced (chapters and cover) so the caller can remove them
   * from storage, or None when the book does not exist.
   */
  def delete(id: Long): Future[Option[List[String]]] = Future {
    db.withTransaction { implicit c =>
      findRow(id).map { book =>
        val chapterKeys = SQL"SELECT content_key FROM book_chapters WHERE book_id = $id".as(str("content_key").*)
        SQL"DELETE FROM books WHERE id = $id".executeUpdate()
        chapterKeys ++ book.coverKey
      }
    }
  }

  /** Points the book at a new cover. Some(previous key) on success, None when the book does not exist. */
  def setCover(id: Long, key: Option[String], contentType: Option[String]): Future[Option[Option[String]]] = Future {
    db.withTransaction { implicit c =>
      findRow(id).map { book =>
        SQL"""UPDATE books SET cover_key = $key, cover_content_type = $contentType, updated_at = CURRENT_TIMESTAMP
              WHERE id = $id""".executeUpdate()
        book.coverKey
      }
    }
  }

  // ---- Chapters ------------------------------------------------------------

  /**
   * A book's chapters in reading order. The cursor is the number of the last
   * chapter returned. None when the book does not exist.
   */
  def chapters(bookId: Long, page: PageRequest): Future[Option[Page[ChapterRef]]] = Future {
    val fetch = page.limit + 1
    db.withConnection { implicit c =>
      if (!bookExists(bookId)) None
      else {
        val after = page.after.getOrElse(0L)
        val rows =
          SQL(s"""SELECT $ChapterColumns FROM book_chapters
                 |WHERE book_id = {bookId} AND chapter_number > {after}
                 |ORDER BY chapter_number LIMIT {fetch}""".stripMargin)
            .on("bookId" -> bookId, "after" -> after, "fetch" -> fetch)
            .as(chapterParser.*)
        Some(Page.fromFetched(rows, page.limit)(_.number.toLong))
      }
    }
  }

  def chapter(bookId: Long, number: Int): Future[Option[ChapterRef]] = Future {
    db.withConnection { implicit c =>
      SQL(s"SELECT $ChapterColumns FROM book_chapters WHERE book_id = {bookId} AND chapter_number = {number}")
        .on("bookId" -> bookId, "number" -> number)
        .as(chapterParser.singleOpt)
    }
  }

  def chapterRefs(bookId: Long): Future[List[ChapterRef]] = Future {
    db.withConnection { implicit c =>
      SQL(s"SELECT $ChapterColumns FROM book_chapters WHERE book_id = {bookId} ORDER BY chapter_number")
        .on("bookId" -> bookId)
        .as(chapterParser.*)
    }
  }

  /**
   * Creates or replaces chapter `ref.number`. Some(true) when it was created,
   * Some(false) when replaced, None when the book does not exist.
   */
  def upsertChapter(ref: ChapterRef): Future[Option[Boolean]] = Future {
    db.withConnection { implicit c =>
      def update(): Boolean =
        SQL"""UPDATE book_chapters SET title = ${ref.title}, content_key = ${ref.contentKey},
                content_type = ${ref.contentType}, size_bytes = ${ref.sizeBytes}, updated_at = CURRENT_TIMESTAMP
              WHERE book_id = ${ref.bookId} AND chapter_number = ${ref.number}""".executeUpdate() > 0

      if (!bookExists(ref.bookId)) None
      else if (update()) Some(false)
      else
        try {
          SQL"""INSERT INTO book_chapters (book_id, chapter_number, title, content_key, content_type, size_bytes)
                VALUES (${ref.bookId}, ${ref.number}, ${ref.title}, ${ref.contentKey}, ${ref.contentType},
                        ${ref.sizeBytes})""".executeUpdate()
          Some(true)
        } catch {
          // A concurrent request created it first: replace theirs instead.
          case e: java.sql.SQLException if SqlUtil.isUniqueViolation(e) => update(); Some(false)
          // The book was deleted in the meantime.
          case e: java.sql.SQLException if e.getSQLState == "23503" || e.getSQLState == "23506" => None
        }
    }
  }

  /** Deletes the chapter and returns its object key, None when it does not exist. */
  def deleteChapter(bookId: Long, number: Int): Future[Option[String]] = Future {
    db.withTransaction { implicit c =>
      val key = SQL"SELECT content_key FROM book_chapters WHERE book_id = $bookId AND chapter_number = $number"
        .as(str("content_key").singleOpt)
      key.foreach(_ => SQL"DELETE FROM book_chapters WHERE book_id = $bookId AND chapter_number = $number".executeUpdate())
      key
    }
  }

  def isReady: Future[Boolean] =
    Future(db.withConnection(implicit c => SQL"SELECT 1".as(scalar[Int].single) == 1)).recover { case _ => false }

  // ---- Helpers -------------------------------------------------------------

  private def findRow(id: Long)(implicit c: Connection): Option[BookRow] =
    SQL(s"SELECT $BookColumns FROM books b WHERE b.id = {id}").on("id" -> id).as(bookParser.singleOpt)

  private def bookExists(id: Long)(implicit c: Connection): Boolean =
    SQL"SELECT 1 FROM books WHERE id = $id".as(scalar[Int].singleOpt).isDefined

  private def linkCategories(bookId: Long, categoryIds: Seq[Long])(implicit c: Connection): Unit =
    categoryIds.distinct.foreach { categoryId =>
      SQL"INSERT INTO book_categories (book_id, category_id) VALUES ($bookId, $categoryId)".executeUpdate()
    }

  private def uniqueIsbn[A](input: BookInput)(block: => A): Either[String, A] =
    try Right(block)
    catch {
      case e: java.sql.SQLException if SqlUtil.isUniqueViolation(e) =>
        Left(s"A book with ISBN ${input.isbn.getOrElse("")} already exists")
    }
}
