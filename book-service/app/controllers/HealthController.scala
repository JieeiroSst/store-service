package controllers

import javax.inject._

import play.api.libs.json.Json
import play.api.mvc._
import services.BookService

import scala.concurrent.ExecutionContext

@Singleton
class HealthController @Inject() (cc: ControllerComponents, books: BookService)(implicit ec: ExecutionContext)
    extends AbstractController(cc) {

  /** Liveness: the process is serving requests. */
  def live: Action[AnyContent] = Action(Ok(Json.obj("status" -> "ok")))

  /** Readiness: the database and the MinIO bucket both answer. */
  def ready: Action[AnyContent] = Action.async {
    books.isReady.map {
      case true  => Ok(Json.obj("status" -> "ready"))
      case false => ServiceUnavailable(Json.obj("status" -> "not ready"))
    }
  }
}
