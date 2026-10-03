package controllers

import javax.inject._

import auth.AdminAuth
import models.{BookFilter, CategoryInput, PageRequest}
import play.api.libs.json._
import play.api.mvc._
import services.{BookService, ServiceError}

import scala.concurrent.{ExecutionContext, Future}

@Singleton
class CategoryController @Inject() (cc: ControllerComponents, service: BookService, adminAuth: AdminAuth)(implicit
    ec: ExecutionContext
) extends AbstractController(cc) {

  private val Admin = Action andThen adminAuth

  /** GET /api/v1/categories */
  def list: Action[AnyContent] = Action.async {
    service.allCategories().map(cs => Ok(Json.obj("data" -> cs)))
  }

  /** GET /api/v1/categories/:slug */
  def show(slug: String): Action[AnyContent] = Action.async {
    service.findCategory(slug).map {
      case Some(c) => Ok(Json.toJson(c))
      case None    => notFound(slug)
    }
  }

  /** GET /api/v1/categories/:slug/books?cursor=&limit=&q=: books in the category, cursor-paginated */
  def books(slug: String, cursor: Option[String], limit: Option[Int], q: Option[String]): Action[AnyContent] =
    Action.async {
      service.findCategory(slug).flatMap {
        case None => Future.successful(notFound(slug))
        case Some(category) =>
          PageRequest.parse(cursor, limit) match {
            case Left(error) => Future.successful(BadRequest(Json.obj("error" -> error)))
            case Right(page) =>
              service
                .list(BookFilter.from(q, None, None, Some(category.slug)), page)
                .map(p => Ok(Json.toJson(p)))
          }
      }
    }

  /** POST /api/v1/categories with {"slug", "name"} */
  def create: Action[JsValue] = Admin.async(parse.json) { request =>
    request.body.validate[CategoryInput] match {
      case errors: JsError =>
        Future.successful(BadRequest(Json.obj("error" -> "Invalid request body", "details" -> JsError.toJson(errors))))
      case JsSuccess(input, _) =>
        service.createCategory(input).map {
          case Right(c) => Created(Json.toJson(c)).withHeaders(LOCATION -> routes.CategoryController.show(c.slug).url)
          case Left(ServiceError.Conflict(m)) => Conflict(Json.obj("error" -> m))
          case Left(e)                        => BadRequest(Json.obj("error" -> e.message))
        }
    }
  }

  /** PUT /api/v1/categories/:slug with {"name"} (the slug is the stable identifier) */
  def rename(slug: String): Action[JsValue] = Admin.async(parse.json) { request =>
    (request.body \ "name").validate[String].map(_.trim).filter(n => n.nonEmpty && n.length <= 128) match {
      case JsSuccess(name, _) =>
        service.renameCategory(slug, name).map {
          case Some(c) => Ok(Json.toJson(c))
          case None    => notFound(slug)
        }
      case _ => Future.successful(BadRequest(Json.obj("error" -> "name is required (max 128 characters)")))
    }
  }

  /** DELETE /api/v1/categories/:slug (books stay, only the link is removed) */
  def delete(slug: String): Action[AnyContent] = Admin.async {
    service.deleteCategory(slug).map(if (_) NoContent else notFound(slug))
  }

  private def notFound(slug: String): Result = NotFound(Json.obj("error" -> s"Category $slug not found"))
}
