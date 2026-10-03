package services

import javax.inject._

import akka.actor.ActorSystem
import akka.pattern.after
import models.{BookInput, CategoryInput, ChapterInput}
import play.api.libs.json._
import play.api.{Configuration, Environment}
import repositories.BookRepository
import storage.ContentStore

import scala.concurrent.duration._
import scala.concurrent.{ExecutionContext, Future}
import scala.util.Using

/**
 * On startup, makes sure the MinIO bucket exists and, when
 * `book.seed.enabled`, loads the sample catalogue from conf/seed/books.json:
 * missing categories and books (matched by ISBN) are created through
 * BookService, and chapter objects missing from MinIO (e.g. after its volume
 * was wiped) are uploaded again. MinIO or the database may come up after
 * this service, so it keeps retrying instead of failing the boot.
 */
@Singleton
class ContentSeeder @Inject() (
    config: Configuration,
    env: Environment,
    service: BookService,
    books: BookRepository,
    store: ContentStore,
    system: ActorSystem
)(implicit ec: ExecutionContext) {
  import ContentSeeder._

  private val logger = org.slf4j.LoggerFactory.getLogger(classOf[ContentSeeder])

  private val MaxAttempts = 30
  private val RetryDelay  = 10.seconds

  /** Number of books created; completes when seeding is done or disabled. */
  val done: Future[Int] =
    if (config.get[Boolean]("book.seed.enabled")) attempt(1)
    else store.ensureBucket().map(_ => 0).recover { case _ => 0 }

  private def attempt(n: Int): Future[Int] =
    seed().recoverWith {
      case e if n < MaxAttempts =>
        logger.warn(s"Seeding books failed (attempt $n/$MaxAttempts): ${e.getMessage}; retrying in $RetryDelay")
        after(RetryDelay, system.scheduler)(attempt(n + 1))
      case e =>
        logger.error("Giving up seeding books", e)
        Future.failed(e)
    }

  private def seed(): Future[Int] = {
    val data = loadSeed()
    for {
      _       <- store.ensureBucket()
      _       <- sequentially(data.categories)(c => service.createCategory(c).map(_ => ()))
      created <- sequentially(data.books)(seedBook) // in order, so ids follow the file
    } yield {
      val count = created.count(identity)
      logger.info(s"Seed catalogue loaded: $count books created, ${created.size - count} already present")
      count
    }
  }

  /** True when the book was created. */
  private def seedBook(book: SeedBook): Future[Boolean] =
    books.findByIsbn(book.isbn).flatMap {
      case Some(existing) => repairChapters(existing.id, book).map(_ => false)
      case None =>
        service.create(book.input).flatMap {
          case Right(detail) =>
            sequentially(book.chapters.zipWithIndex) { case (c, i) =>
              service.putChapter(detail.book.id, i + 1, c).map(_ => ())
            }.map(_ => true)
          // Another replica created it at the same time.
          case Left(_: ServiceError.Conflict) => Future.successful(false)
          case Left(error)                    => Future.failed(new IllegalStateException(error.message))
        }
    }

  private def repairChapters(bookId: Long, book: SeedBook): Future[Unit] =
    books.chapterRefs(bookId).flatMap { refs =>
      sequentially(refs) { ref =>
        store.exists(ref.contentKey).flatMap {
          case true => Future.unit
          case false =>
            book.chapters.lift(ref.number - 1) match {
              case Some(c) =>
                logger.info(s"Re-uploading missing object ${ref.contentKey}")
                store.put(ref.contentKey, c.content, ref.contentType)
              case None => Future.unit
            }
        }
      }.map(_ => ())
    }

  private def loadSeed(): SeedData =
    env.resourceAsStream("seed/books.json") match {
      case None     => SeedData(Nil, Nil)
      case Some(in) => Using.resource(in)(Json.parse).as[SeedData]
    }

  private def sequentially[A, B](items: Seq[A])(f: A => Future[B]): Future[Seq[B]] =
    items.foldLeft(Future.successful(Vector.empty[B])) { (acc, item) =>
      acc.flatMap(done => f(item).map(done :+ _))
    }
}

object ContentSeeder {
  final case class SeedBook(
      title: String,
      author: String,
      isbn: String,
      description: Option[String],
      publishedYear: Option[Int],
      categories: Seq[String],
      chapters: Seq[ChapterInput]
  ) {
    def input: BookInput = BookInput(title, author, Some(isbn), description, publishedYear, categories)
  }

  final case class SeedData(categories: Seq[CategoryInput], books: Seq[SeedBook])

  implicit val seedBookReads: Reads[SeedBook] = Json.reads[SeedBook]
  implicit val seedDataReads: Reads[SeedData] = Json.reads[SeedData]
}
