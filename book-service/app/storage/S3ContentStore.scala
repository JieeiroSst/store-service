package storage

import java.net.URI
import java.time.Duration
import javax.inject._

import akka.actor.ActorSystem
import play.api.Configuration
import play.api.inject.ApplicationLifecycle
import play.api.libs.concurrent.CustomExecutionContext
import software.amazon.awssdk.auth.credentials.{AwsBasicCredentials, StaticCredentialsProvider}
import software.amazon.awssdk.core.sync.RequestBody
import software.amazon.awssdk.http.urlconnection.UrlConnectionHttpClient
import software.amazon.awssdk.regions.Region
import software.amazon.awssdk.services.s3.model._
import software.amazon.awssdk.services.s3.presigner.S3Presigner
import software.amazon.awssdk.services.s3.presigner.model.GetObjectPresignRequest
import software.amazon.awssdk.services.s3.{S3Client, S3Configuration}

import scala.concurrent.Future

/** Thread pool for blocking S3 calls, configured as `storage.dispatcher`. */
@Singleton
class StorageExecutionContext @Inject() (system: ActorSystem)
    extends CustomExecutionContext(system, "storage.dispatcher")

/**
 * ContentStore on any S3-compatible server. Path-style addressing
 * (`http://minio-svc:9000/<bucket>/<key>`) is what MinIO expects.
 */
@Singleton
class S3ContentStore @Inject() (config: Configuration, lifecycle: ApplicationLifecycle)(implicit
    ec: StorageExecutionContext
) extends ContentStore {

  private val bucket = config.get[String]("storage.bucket")
  private val region = Region.of(config.get[String]("storage.region"))
  private val credentials = StaticCredentialsProvider.create(
    AwsBasicCredentials.create(config.get[String]("storage.accessKey"), config.get[String]("storage.secretKey"))
  )
  private val pathStyle = S3Configuration.builder().pathStyleAccessEnabled(true).build()

  private val client: S3Client = S3Client
    .builder()
    .endpointOverride(URI.create(config.get[String]("storage.endpoint")))
    .region(region)
    .credentialsProvider(credentials)
    .serviceConfiguration(pathStyle)
    .httpClientBuilder(UrlConnectionHttpClient.builder())
    .build()

  // The in-cluster endpoint (minio-svc) is not reachable by browsers, so
  // presigned URLs are signed for the public one when it is configured.
  private val presignTtl = Duration.ofMillis(config.get[scala.concurrent.duration.FiniteDuration]("storage.presignTtl").toMillis)
  private val presigner: Option[S3Presigner] =
    config.getOptional[String]("storage.publicEndpoint").map(_.trim).filter(_.nonEmpty).map { endpoint =>
      S3Presigner
        .builder()
        .endpointOverride(URI.create(endpoint))
        .region(region)
        .credentialsProvider(credentials)
        .serviceConfiguration(pathStyle)
        .build()
    }

  lifecycle.addStopHook { () =>
    presigner.foreach(_.close())
    Future.successful(client.close())
  }

  override def getBytes(key: String): Future[Option[StoredObject]] = Future {
    try {
      val obj = client.getObjectAsBytes(GetObjectRequest.builder().bucket(bucket).key(key).build())
      Some(StoredObject(obj.asByteArray(), Option(obj.response().contentType()).getOrElse("application/octet-stream")))
    } catch { case e: S3Exception if e.statusCode == 404 => None }
  }

  override def putBytes(key: String, bytes: Array[Byte], contentType: String): Future[Unit] = Future {
    client.putObject(
      PutObjectRequest.builder().bucket(bucket).key(key).contentType(contentType).build(),
      RequestBody.fromBytes(bytes)
    )
    ()
  }

  override def delete(key: String): Future[Unit] = Future {
    client.deleteObject(DeleteObjectRequest.builder().bucket(bucket).key(key).build())
    ()
  }

  override def exists(key: String): Future[Boolean] = Future {
    try { client.headObject(HeadObjectRequest.builder().bucket(bucket).key(key).build()); true }
    catch { case e: S3Exception if e.statusCode == 404 => false }
  }

  override def presignedGetUrl(key: String): Option[String] =
    presigner.map { p =>
      p.presignGetObject(
        GetObjectPresignRequest
          .builder()
          .signatureDuration(presignTtl)
          .getObjectRequest(GetObjectRequest.builder().bucket(bucket).key(key).build())
          .build()
      ).url()
        .toString
    }

  override def ensureBucket(): Future[Unit] = Future {
    try { client.headBucket(HeadBucketRequest.builder().bucket(bucket).build()); () }
    catch {
      case e: S3Exception if e.statusCode == 404 =>
        try { client.createBucket(CreateBucketRequest.builder().bucket(bucket).build()); () }
        catch { case _: BucketAlreadyOwnedByYouException | _: BucketAlreadyExistsException => () }
    }
  }

  override def isReady: Future[Boolean] =
    Future { client.headBucket(HeadBucketRequest.builder().bucket(bucket).build()); true }
      .recover { case _ => false }
}
