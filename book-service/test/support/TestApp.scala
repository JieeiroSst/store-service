package support

import java.util.UUID

import org.scalatest.TestSuite
import org.scalatestplus.play.guice.GuiceFakeApplicationFactory
import play.api.Application
import play.api.inject.bind
import play.api.inject.guice.GuiceApplicationBuilder
import storage.ContentStore

/**
 * Builds the application with an in-memory ContentStore instead of MinIO and
 * its own in-memory database, so suites running in parallel don't share state.
 */
trait TestApp extends GuiceFakeApplicationFactory { this: TestSuite =>
  val AdminKey = "test-admin-key"

  lazy val contentStore = new InMemoryContentStore

  override def fakeApplication(): Application =
    new GuiceApplicationBuilder()
      .configure(
        "db.default.url"    -> s"jdbc:h2:mem:${UUID.randomUUID()};MODE=PostgreSQL;DATABASE_TO_LOWER=TRUE",
        "book.admin.apiKey" -> AdminKey
      )
      .overrides(bind[ContentStore].toInstance(contentStore))
      .build()
}
