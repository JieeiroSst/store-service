package storage

import java.nio.charset.StandardCharsets

import scala.concurrent.{ExecutionContext, Future}

final case class StoredObject(bytes: Array[Byte], contentType: String)

/** Object storage holding chapter text and cover images (MinIO / S3 in practice). */
trait ContentStore {

  /** None when no object exists under `key`. */
  def getBytes(key: String): Future[Option[StoredObject]]

  def putBytes(key: String, bytes: Array[Byte], contentType: String): Future[Unit]

  /** Idempotent: deleting a missing object succeeds. */
  def delete(key: String): Future[Unit]

  def exists(key: String): Future[Boolean]

  /**
   * A time-limited URL clients can download the object from directly, or None
   * when storage has no publicly reachable endpoint configured.
   */
  def presignedGetUrl(key: String): Option[String]

  /** Creates the bucket if it does not exist yet. */
  def ensureBucket(): Future[Unit]

  /** True when the bucket is reachable; used by the readiness probe. */
  def isReady: Future[Boolean]

  final def get(key: String)(implicit ec: ExecutionContext): Future[Option[String]] =
    getBytes(key).map(_.map(o => new String(o.bytes, StandardCharsets.UTF_8)))

  final def put(key: String, content: String, contentType: String): Future[Unit] =
    putBytes(key, content.getBytes(StandardCharsets.UTF_8), contentType)
}
