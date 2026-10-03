package services

import java.nio.charset.StandardCharsets
import javax.inject._

import models._
import repositories.{BookRepository, CategoryRepository}
import storage.ContentStore

import scala.concurrent.{ExecutionContext, Future}

/** A chapter row whose text object is missing from object storage. */
final case class ContentUnavailableException(keys: Seq[String])
    extends RuntimeException(s"Content not found in object storage: ${keys.mkString(", ")}")

/** Why a write was refused; the controller maps each case to a status code. */
sealed trait ServiceError { def message: String }
object ServiceError {
  final case class NotFound(message: String) extends ServiceError
  final case class Invalid(message: String)  extends ServiceError
  final case class Conflict(message: String) extends ServiceError
}

/**
 * Joins book metadata (database) with chapter text and covers (object
 * storage). Writes put the object first and the row second, so a row never
 * points at an object that was not stored. Deletes go the other way: row
 * first, then a best-effort object delete.
 */
@Singleton
class BookService @Inject() (books: BookRepository, categories: CategoryRepository, store: ContentStore)(implicit
    ec: ExecutionContext
) {
  import ServiceError._

  private val logger = org.slf4j.LoggerFactory.getLogger(classOf[BookService])

  val CoverContentTypes: Set[String] = Set("image/jpeg", "image/png", "image/webp", "image/gif")
  val MaxCoverBytes: Int             = 5 * 1024 * 1024

  // ---- Books ---------------------------------------------------------------

  def list(filter: BookFilter, page: PageRequest): Future[Page[Book]] =
    for {
      rows <- books.list(filter, page)
      cats <- books.categoriesFor(rows.items.map(_.id))
    } yield rows.copy(items = rows.items.map(r => toBook(r, cats.getOrElse(r.id, Nil))))

  def find(id: Long): Future[Option[BookDetail]] =
    books.find(id).flatMap {
      case None => Future.successful(None)
      case Some((row, chapterCount)) =>
        books.categoriesFor(Seq(id)).map(cats => Some(BookDetail(toBook(row, cats.getOrElse(id, Nil)), chapterCount)))
    }

  def create(input: BookInput): Future[Either[ServiceError, BookDetail]] =
    withCategories(input.categories) { categoryIds =>
      books.insert(input, categoryIds).flatMap {
        case Left(conflict) => Future.successful(Left(Conflict(conflict)))
        case Right(id)      => found(id)
      }
    }

  def update(id: Long, input: BookInput): Future[Either[ServiceError, BookDetail]] =
    withCategories(input.categories) { categoryIds =>
      books.update(id, input, categoryIds).flatMap {
        case Left(conflict) => Future.successful(Left(Conflict(conflict)))
        case Right(false)   => Future.successful(Left(NotFound(s"Book $id not found")))
        case Right(true)    => found(id)
      }
    }

  /** False when the book does not exist. */
  def delete(id: Long): Future[Boolean] =
    books.delete(id).map {
      case None => false
      case Some(keys) =>
        keys.foreach(deleteObject)
        true
    }

  // ---- Covers --------------------------------------------------------------

  def putCover(bookId: Long, bytes: Array[Byte], contentType: String): Future[Either[ServiceError, BookDetail]] =
    if (!CoverContentTypes.contains(contentType))
      Future.successful(Left(Invalid(s"Cover must be one of ${CoverContentTypes.toSeq.sorted.mkString(", ")}")))
    else if (bytes.isEmpty) Future.successful(Left(Invalid("Cover image is empty")))
    else
      books.find(bookId).flatMap {
        case None => Future.successful(Left(NotFound(s"Book $bookId not found")))
        case Some(_) =>
          // A new key per upload: CDNs and browsers never serve a stale cover
          // from a cached presigned URL.
          val key = s"books/$bookId/cover-${System.currentTimeMillis()}"
          for {
            _        <- store.putBytes(key, bytes, contentType)
            previous <- books.setCover(bookId, Some(key), Some(contentType))
            _ = previous.flatten.foreach(deleteObject)
            result <- found(bookId)
          } yield result
      }

  /** False when the book does not exist. */
  def deleteCover(bookId: Long): Future[Boolean] =
    books.setCover(bookId, None, None).map {
      case None => false
      case Some(previous) =>
        previous.foreach(deleteObject)
        true
    }

  /** The cover bytes, for clients that cannot use a presigned URL. */
  def cover(bookId: Long): Future[Option[storage.StoredObject]] =
    books.find(bookId).flatMap {
      case Some((row, _)) if row.coverKey.isDefined =>
        store.getBytes(row.coverKey.get).map(_.map(o => o.copy(contentType = row.coverContentType.getOrElse(o.contentType))))
      case _ => Future.successful(None)
    }

  // ---- Chapters ------------------------------------------------------------

  /** None when the book does not exist. Chapter texts are fetched concurrently. */
  def chapters(bookId: Long, page: PageRequest): Future[Option[Page[Chapter]]] =
    books.chapters(bookId, page).flatMap {
      case None => Future.successful(None)
      case Some(refs) =>
        Future.traverse(refs.items)(ref => store.get(ref.contentKey).map(ref -> _)).map { loaded =>
          val missing = loaded.collect { case (ref, None) => ref.contentKey }
          if (missing.nonEmpty) throw ContentUnavailableException(missing)
          Some(refs.copy(items = loaded.collect { case (ref, Some(text)) => toChapter(ref, text) }))
        }
    }

  def chapter(bookId: Long, number: Int): Future[Option[Chapter]] =
    books.chapter(bookId, number).flatMap {
      case None => Future.successful(None)
      case Some(ref) =>
        store.get(ref.contentKey).map {
          case Some(text) => Some(toChapter(ref, text))
          case None       => throw ContentUnavailableException(Seq(ref.contentKey))
        }
    }

  /** Right(true) when the chapter was created, Right(false) when replaced. */
  def putChapter(bookId: Long, number: Int, input: ChapterInput): Future[Either[ServiceError, (Boolean, Chapter)]] =
    if (number < 1) Future.successful(Left(Invalid("Chapter number must be 1 or greater")))
    else
      books.find(bookId).flatMap {
        case None => Future.successful(Left(NotFound(s"Book $bookId not found")))
        case Some(_) =>
          val ref = ChapterRef(
            bookId,
            number,
            input.title,
            chapterKey(bookId, number),
            "text/plain; charset=utf-8",
            Some(input.content.getBytes(StandardCharsets.UTF_8).length.toLong)
          )
          for {
            _       <- store.put(ref.contentKey, input.content, ref.contentType)
            created <- books.upsertChapter(ref)
          } yield created match {
            case Some(isNew) => Right(isNew -> toChapter(ref, input.content))
            case None        => Left(NotFound(s"Book $bookId not found"))
          }
      }

  /** False when the chapter does not exist. */
  def deleteChapter(bookId: Long, number: Int): Future[Boolean] =
    books.deleteChapter(bookId, number).map {
      case None => false
      case Some(key) =>
        deleteObject(key)
        true
    }

  // ---- Categories ----------------------------------------------------------

  def allCategories(): Future[List[Category]] = categories.all()

  def findCategory(slug: String): Future[Option[Category]] = categories.find(slug.toLowerCase)

  def createCategory(input: CategoryInput): Future[Either[ServiceError, Category]] =
    categories.insert(input).map(_.left.map(Conflict))

  def renameCategory(slug: String, name: String): Future[Option[Category]] = categories.rename(slug.toLowerCase, name)

  def deleteCategory(slug: String): Future[Boolean] = categories.delete(slug.toLowerCase)

  // ---- Misc ----------------------------------------------------------------

  def isReady: Future[Boolean] = books.isReady.zipWith(store.isReady)(_ && _)

  def chapterKey(bookId: Long, number: Int): String = s"books/$bookId/chapters/$number.txt"

  private def found(id: Long): Future[Either[ServiceError, BookDetail]] =
    find(id).map(_.toRight(NotFound(s"Book $id not found")))

  /** Resolves category slugs to ids; unknown slugs are a validation error. */
  private def withCategories[A](slugs: Seq[String])(
      f: Seq[Long] => Future[Either[ServiceError, A]]
  ): Future[Either[ServiceError, A]] =
    categories.findBySlugs(slugs).flatMap { found =>
      val unknown = slugs.filterNot(found.map(_.slug).toSet)
      if (unknown.nonEmpty) Future.successful(Left(Invalid(s"Unknown categories: ${unknown.mkString(", ")}")))
      else f(found.map(_.id))
    }

  private def deleteObject(key: String): Unit =
    store.delete(key).failed.foreach(e => logger.warn(s"Could not delete object $key from storage: ${e.getMessage}"))

  private def coverUrl(row: BookRow): Option[String] =
    row.coverKey.map(key => store.presignedGetUrl(key).getOrElse(s"/api/v1/books/${row.id}/cover"))

  private def toBook(row: BookRow, cats: Seq[Category]): Book =
    Book(row.id, row.title, row.author, row.isbn, row.description, row.publishedYear, cats, coverUrl(row),
      row.createdAt, row.updatedAt)

  private def toChapter(ref: ChapterRef, text: String): Chapter =
    Chapter(ref.bookId, ref.number, ref.title, ref.contentType, ref.sizeBytes, text)
}
