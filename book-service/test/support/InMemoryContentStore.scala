package support

import java.util.concurrent.ConcurrentHashMap

import storage.{ContentStore, StoredObject}

import scala.concurrent.Future

/** ContentStore kept in memory, standing in for MinIO in tests. */
class InMemoryContentStore extends ContentStore {
  val objects = new ConcurrentHashMap[String, StoredObject]()

  def text(key: String): Option[String] = Option(objects.get(key)).map(o => new String(o.bytes, "UTF-8"))

  override def getBytes(key: String): Future[Option[StoredObject]] = Future.successful(Option(objects.get(key)))

  override def putBytes(key: String, bytes: Array[Byte], contentType: String): Future[Unit] = {
    objects.put(key, StoredObject(bytes, contentType))
    Future.unit
  }

  override def delete(key: String): Future[Unit] = {
    objects.remove(key)
    Future.unit
  }

  override def exists(key: String): Future[Boolean] = Future.successful(objects.containsKey(key))
  override def presignedGetUrl(key: String): Option[String] = None
  override def ensureBucket(): Future[Unit] = Future.unit
  override def isReady: Future[Boolean] = Future.successful(true)
}
