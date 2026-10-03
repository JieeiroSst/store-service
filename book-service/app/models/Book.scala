package models

import java.time.Instant

import play.api.libs.functional.syntax._
import play.api.libs.json._

final case class Category(id: Long, slug: String, name: String)

object Category {
  implicit val writes: OWrites[Category] = OWrites(c => Json.obj("slug" -> c.slug, "name" -> c.name))
}

/** A row of `books`; the cover is referenced by its object key in MinIO. */
final case class BookRow(
    id: Long,
    title: String,
    author: String,
    isbn: Option[String],
    description: Option[String],
    publishedYear: Option[Int],
    coverKey: Option[String],
    coverContentType: Option[String],
    createdAt: Instant,
    updatedAt: Instant
)

/** A book as the API returns it. */
final case class Book(
    id: Long,
    title: String,
    author: String,
    isbn: Option[String],
    description: Option[String],
    publishedYear: Option[Int],
    categories: Seq[Category],
    coverUrl: Option[String],
    createdAt: Instant,
    updatedAt: Instant
)

object Book {
  implicit val writes: OWrites[Book] = Json.writes[Book]
}

final case class BookDetail(book: Book, chapterCount: Int)

object BookDetail {
  implicit val writes: OWrites[BookDetail] = OWrites { d =>
    Json.toJsObject(d.book) + ("chapterCount" -> JsNumber(d.chapterCount))
  }
}

/** A chapter row: metadata plus where its text lives in object storage. */
final case class ChapterRef(
    bookId: Long,
    number: Int,
    title: String,
    contentKey: String,
    contentType: String,
    sizeBytes: Option[Long]
)

final case class Chapter(
    bookId: Long,
    number: Int,
    title: String,
    contentType: String,
    sizeBytes: Option[Long],
    content: String
)

object Chapter {
  implicit val writes: OWrites[Chapter] = Json.writes[Chapter]
}

/** Filters for the book list; all optional and combined with AND. */
final case class BookFilter(
    q: Option[String] = None,
    author: Option[String] = None,
    year: Option[Int] = None,
    category: Option[String] = None
)

object BookFilter {
  /** Trims inputs and drops empty ones (`?q=` means no filter). */
  def from(q: Option[String], author: Option[String], year: Option[Int], category: Option[String]): BookFilter = {
    def clean(s: Option[String]) = s.map(_.trim).filter(_.nonEmpty)
    new BookFilter(clean(q), clean(author), year, clean(category).map(_.toLowerCase))
  }
}

// ---- Request bodies -------------------------------------------------------

private[models] object Validation {
  val SlugPattern = "^[a-z0-9]+(?:-[a-z0-9]+)*$".r

  def text(max: Int): Reads[String] =
    Reads.of[String]
      .map(_.trim)
      .filter(JsonValidationError("error.required"))(_.nonEmpty)
      .filter(JsonValidationError("error.maxLength", max))(_.length <= max)

  def optionalText(max: Int): Reads[String] =
    Reads.of[String]
      .map(_.trim)
      .filter(JsonValidationError("error.maxLength", max))(_.length <= max)

  /** ISBN-10 or ISBN-13, hyphens and spaces stripped. */
  val isbn: Reads[String] =
    Reads.of[String]
      .map(_.replaceAll("[\\s-]", "").toUpperCase)
      .filter(JsonValidationError("error.isbn"))(s => s.matches("^\\d{9}[\\dX]$") || s.matches("^\\d{13}$"))

  val slug: Reads[String] =
    Reads.of[String]
      .map(_.trim.toLowerCase)
      .filter(JsonValidationError("error.slug"))(s => s.length <= 64 && SlugPattern.matches(s))
}

final case class BookInput(
    title: String,
    author: String,
    isbn: Option[String],
    description: Option[String],
    publishedYear: Option[Int],
    categories: Seq[String]
)

object BookInput {
  import Validation._

  implicit val reads: Reads[BookInput] = (
    (__ \ "title").read(text(255)) and
      (__ \ "author").read(text(255)) and
      (__ \ "isbn").readNullable(isbn) and
      (__ \ "description").readNullable(optionalText(4000)).map(_.filter(_.nonEmpty)) and
      (__ \ "publishedYear").readNullable(Reads.min(0) keepAnd Reads.max(2100)) and
      (__ \ "categories").readWithDefault[Seq[String]](Nil)(Reads.seq(slug)).map(_.distinct)
  )(BookInput.apply _)
}

final case class ChapterInput(title: String, content: String)

object ChapterInput {
  import Validation._

  /** 1 MiB of text per chapter; split longer chapters. */
  val MaxContentLength: Int = 1024 * 1024

  implicit val reads: Reads[ChapterInput] = (
    (__ \ "title").read(text(255)) and
      (__ \ "content").read(
        Reads.of[String]
          .filter(JsonValidationError("error.required"))(_.trim.nonEmpty)
          .filter(JsonValidationError("error.maxLength", MaxContentLength))(_.length <= MaxContentLength)
      )
  )(ChapterInput.apply _)
}

final case class CategoryInput(slug: String, name: String)

object CategoryInput {
  import Validation._

  implicit val reads: Reads[CategoryInput] = (
    (__ \ "slug").read(slug) and
      (__ \ "name").read(text(128))
  )(CategoryInput.apply _)
}
