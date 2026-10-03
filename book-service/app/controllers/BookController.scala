package controllers

import javax.inject._

import akka.util.ByteString
import auth.AdminAuth
import models._
import play.api.libs.json._
import play.api.mvc._
import services.{BookService, ContentUnavailableException, ServiceError}

import scala.concurrent.{ExecutionContext, Future}

@Singleton
class BookController @Inject() (cc: ControllerComponents, books: BookService, adminAuth: AdminAuth)(implicit
    ec: ExecutionContext
) extends AbstractController(cc) {

  /** Write endpoints: API key required when one is configured. */
  private val Admin = Action andThen adminAuth

  // ---- Books ---------------------------------------------------------------

  /** GET /api/v1/books?cursor=&limit=&q=&author=&year=&category= */
  def list(
      cursor: Option[String],
      limit: Option[Int],
      q: Option[String],
      author: Option[String],
      year: Option[Int],
      category: Option[String]
  ): Action[AnyContent] = Action.async {
    withPage(cursor, limit) { page =>
      books.list(BookFilter.from(q, author, year, category), page).map(p => Ok(Json.toJson(p)))
    }
  }

  /** GET /api/v1/books/:id */
  def show(id: Long): Action[AnyContent] = Action.async {
    books.find(id).map {
      case Some(book) => Ok(Json.toJson(book))
      case None       => notFound(s"Book $id not found")
    }
  }

  /** POST /api/v1/books */
  def create: Action[JsValue] = Admin.async(parse.json) { request =>
    withJson[BookInput](request) { input =>
      books.create(input).map {
        case Right(book) => Created(Json.toJson(book)).withHeaders(LOCATION -> routes.BookController.show(book.book.id).url)
        case Left(error) => errorResult(error)
      }
    }
  }

  /** PUT /api/v1/books/:id (replaces every field, including categories) */
  def update(id: Long): Action[JsValue] = Admin.async(parse.json) { request =>
    withJson[BookInput](request) { input =>
      books.update(id, input).map {
        case Right(book) => Ok(Json.toJson(book))
        case Left(error) => errorResult(error)
      }
    }
  }

  /** DELETE /api/v1/books/:id (also removes its chapters and cover from storage) */
  def delete(id: Long): Action[AnyContent] = Admin.async {
    books.delete(id).map(if (_) NoContent else notFound(s"Book $id not found"))
  }

  // ---- Cover ---------------------------------------------------------------

  /** GET /api/v1/books/:id/cover (proxied from MinIO when no presigned URL is available) */
  def cover(id: Long): Action[AnyContent] = Action.async {
    books.cover(id).map {
      case Some(obj) => Ok(ByteString(obj.bytes)).as(obj.contentType).withHeaders(CACHE_CONTROL -> "public, max-age=300")
      case None      => notFound(s"Book $id has no cover")
    }
  }

  /** PUT /api/v1/books/:id/cover with the image as the raw body (Content-Type: image/png, ...) */
  def putCover(id: Long): Action[ByteString] = Admin.async(parse.byteString(books.MaxCoverBytes.toLong)) { request =>
    val contentType = request.contentType.getOrElse("").toLowerCase
    books.putCover(id, request.body.toArray, contentType).map {
      case Right(book) => Ok(Json.toJson(book))
      case Left(error) => errorResult(error)
    }
  }

  /** DELETE /api/v1/books/:id/cover */
  def deleteCover(id: Long): Action[AnyContent] = Admin.async {
    books.deleteCover(id).map(if (_) NoContent else notFound(s"Book $id not found"))
  }

  // ---- Chapters ------------------------------------------------------------

  /** GET /api/v1/books/:id/chapters?cursor=&limit= */
  def chapters(id: Long, cursor: Option[String], limit: Option[Int]): Action[AnyContent] = Action.async {
    withPage(cursor, limit) { page =>
      books.chapters(id, page).map {
        case Some(p) => Ok(Json.toJson(p))
        case None    => notFound(s"Book $id not found")
      }.recover(contentUnavailable)
    }
  }

  /** GET /api/v1/books/:id/chapters/:number */
  def chapter(id: Long, number: Int): Action[AnyContent] = Action.async {
    books.chapter(id, number).map {
      case Some(c) => Ok(Json.toJson(c))
      case None    => notFound(s"Chapter $number of book $id not found")
    }.recover(contentUnavailable)
  }

  /** PUT /api/v1/books/:id/chapters/:number with {"title", "content"}: creates or replaces it */
  def putChapter(id: Long, number: Int): Action[JsValue] =
    Admin.async(parse.json(maxLength = 4L * ChapterInput.MaxContentLength)) { request =>
      withJson[ChapterInput](request) { input =>
        books.putChapter(id, number, input).map {
          case Right((true, c))  => Created(Json.toJson(c))
          case Right((false, c)) => Ok(Json.toJson(c))
          case Left(error)       => errorResult(error)
        }
      }
    }

  /** DELETE /api/v1/books/:id/chapters/:number */
  def deleteChapter(id: Long, number: Int): Action[AnyContent] = Admin.async {
    books.deleteChapter(id, number).map(if (_) NoContent else notFound(s"Chapter $number of book $id not found"))
  }

  // ---- Helpers -------------------------------------------------------------

  private def withPage(cursor: Option[String], limit: Option[Int])(f: PageRequest => Future[Result]): Future[Result] =
    PageRequest.parse(cursor, limit) match {
      case Right(page) => f(page)
      case Left(error) => Future.successful(BadRequest(Json.obj("error" -> error)))
    }

  private def withJson[A: Reads](request: Request[JsValue])(f: A => Future[Result]): Future[Result] =
    request.body.validate[A] match {
      case JsSuccess(input, _) => f(input)
      case errors: JsError =>
        Future.successful(BadRequest(Json.obj("error" -> "Invalid request body", "details" -> JsError.toJson(errors))))
    }

  private def notFound(message: String): Result = NotFound(Json.obj("error" -> message))

  private def errorResult(error: ServiceError): Result = error match {
    case ServiceError.NotFound(m) => NotFound(Json.obj("error" -> m))
    case ServiceError.Invalid(m)  => BadRequest(Json.obj("error" -> m))
    case ServiceError.Conflict(m) => Conflict(Json.obj("error" -> m))
  }

  // The row exists but its text is not in object storage (yet).
  private val contentUnavailable: PartialFunction[Throwable, Result] = { case _: ContentUnavailableException =>
    ServiceUnavailable(Json.obj("error" -> "Book content is temporarily unavailable"))
  }
}
