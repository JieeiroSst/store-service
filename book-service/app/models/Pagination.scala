package models

import java.nio.charset.StandardCharsets
import java.util.Base64

import play.api.libs.json._

import scala.util.Try

/**
 * Opaque cursor for keyset pagination. Clients must treat it as an
 * opaque string; internally it wraps the sort key of the last item
 * returned on the previous page.
 */
object Cursor {
  private val Prefix = "v1:"

  def encode(key: Long): String =
    Base64.getUrlEncoder.withoutPadding.encodeToString(s"$Prefix$key".getBytes(StandardCharsets.UTF_8))

  def decode(cursor: String): Option[Long] =
    Try(new String(Base64.getUrlDecoder.decode(cursor), StandardCharsets.UTF_8)).toOption
      .filter(_.startsWith(Prefix))
      .flatMap(s => s.drop(Prefix.length).toLongOption)
      .filter(_ >= 0)
}

final case class PageRequest(after: Option[Long], limit: Int)

object PageRequest {
  val DefaultLimit = 20
  val MaxLimit     = 100

  /** Left(error message) when the cursor cannot be decoded. */
  def parse(cursor: Option[String], limit: Option[Int]): Either[String, PageRequest] = {
    val size = limit.getOrElse(DefaultLimit).max(1).min(MaxLimit)
    cursor.filter(_.nonEmpty) match {
      case None    => Right(PageRequest(None, size))
      case Some(c) => Cursor.decode(c).map(k => PageRequest(Some(k), size)).toRight("Invalid cursor")
    }
  }
}

final case class Page[A](items: Seq[A], limit: Int, nextCursor: Option[String]) {
  def hasMore: Boolean = nextCursor.isDefined
}

object Page {

  /**
   * Builds a page from a query that fetched `limit + 1` rows: the extra
   * row only tells us whether another page exists and is never returned.
   */
  def fromFetched[A](fetched: Seq[A], limit: Int)(key: A => Long): Page[A] = {
    val items = fetched.take(limit)
    val next  = if (fetched.sizeIs > limit) items.lastOption.map(a => Cursor.encode(key(a))) else None
    Page(items, limit, next)
  }

  implicit def writes[A: Writes]: OWrites[Page[A]] = OWrites { page =>
    Json.obj(
      "data" -> page.items,
      "pagination" -> Json.obj(
        "limit"      -> page.limit,
        "nextCursor" -> page.nextCursor,
        "hasMore"    -> page.hasMore
      )
    )
  }
}
