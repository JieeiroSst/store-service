package auth

import java.nio.charset.StandardCharsets
import java.security.MessageDigest
import javax.inject._

import play.api.Configuration
import play.api.libs.json.Json
import play.api.mvc._

import scala.concurrent.{ExecutionContext, Future}

/**
 * Guards the write endpoints with a shared API key in the `X-Api-Key`
 * header (`book.admin.apiKey`, env BOOK_ADMIN_API_KEY). With no key
 * configured, as in `sbt run`, writes are open.
 */
@Singleton
class AdminAuth @Inject() (config: Configuration)(implicit val executionContext: ExecutionContext)
    extends ActionFilter[Request] {

  private val apiKey: Option[Array[Byte]] =
    config.getOptional[String]("book.admin.apiKey").map(_.trim).filter(_.nonEmpty).map(_.getBytes(StandardCharsets.UTF_8))

  override protected def filter[A](request: Request[A]): Future[Option[Result]] = Future.successful {
    apiKey.flatMap { expected =>
      val provided = request.headers.get("X-Api-Key").getOrElse("").getBytes(StandardCharsets.UTF_8)
      if (MessageDigest.isEqual(expected, provided)) None
      else Some(Results.Unauthorized(Json.obj("error" -> "Missing or invalid X-Api-Key")))
    }
  }
}
