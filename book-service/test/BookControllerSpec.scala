import models.{Cursor, PageRequest}
import org.scalatestplus.play.PlaySpec
import org.scalatestplus.play.guice.GuiceOneAppPerSuite
import play.api.libs.json._
import play.api.test.FakeRequest
import play.api.test.Helpers._
import services.ContentSeeder

import scala.concurrent.Await
import scala.concurrent.duration._

/**
 * Runs against the in-memory H2 database seeded by conf/evolutions, with
 * chapter text uploaded by ContentSeeder into an in-memory ContentStore.
 */
class BookControllerSpec extends PlaySpec with GuiceOneAppPerSuite with support.TestApp {

  private lazy val seeded = Await.result(app.injector.instanceOf[ContentSeeder].done, 10.seconds)

  private def get(path: String) = {
    seeded
    route(app, FakeRequest(GET, path)).get
  }

  private def getJson(path: String): JsValue = {
    val result = get(path)
    status(result) mustBe OK
    contentAsJson(result)
  }

  /** Follows nextCursor until hasMore is false, returning every page. */
  private def ids(path: String): Seq[Long] =
    (getJson(path) \ "data").as[Seq[JsObject]].map(b => (b \ "id").as[Long])

  private def walk(basePath: String, limit: Int, query: String = ""): Seq[JsValue] = {
    def loop(cursor: Option[String], acc: Vector[JsValue]): Vector[JsValue] = {
      val path = basePath + s"?limit=$limit" + query + cursor.fold("")(c => s"&cursor=$c")
      val page = getJson(path)
      val next = (page \ "pagination" \ "nextCursor").asOpt[String]
      (page \ "pagination" \ "hasMore").as[Boolean] mustBe next.isDefined
      if (next.isDefined) loop(next, acc :+ page) else acc :+ page
    }
    loop(None, Vector.empty)
  }

  "GET /api/v1/books" should {

    "return the first page newest first" in {
      val page = getJson("/api/v1/books?limit=5")
      (page \ "data").as[Seq[JsObject]].map(b => (b \ "id").as[Long]) mustBe Seq(12L, 11L, 10L, 9L, 8L)
      (page \ "pagination" \ "limit").as[Int] mustBe 5
      (page \ "pagination" \ "hasMore").as[Boolean] mustBe true
    }

    "visit every book exactly once when following cursors" in {
      val pages = walk("/api/v1/books", 5)
      pages.map(p => (p \ "data").as[Seq[JsObject]].size) mustBe Seq(5, 5, 2)
      val ids = pages.flatMap(p => (p \ "data").as[Seq[JsObject]].map(b => (b \ "id").as[Long]))
      ids mustBe (12L to 1L by -1L)
    }

    "not report a next page when the last page is exactly full" in {
      val pages = walk("/api/v1/books", 6)
      pages.map(p => (p \ "data").as[Seq[JsObject]].size) mustBe Seq(6, 6)
    }

    "use the default limit and cap it at the maximum" in {
      (getJson("/api/v1/books") \ "pagination" \ "limit").as[Int] mustBe PageRequest.DefaultLimit
      (getJson("/api/v1/books?limit=100000") \ "pagination" \ "limit").as[Int] mustBe PageRequest.MaxLimit
    }

    "reject an invalid cursor" in {
      status(get("/api/v1/books?cursor=not-a-cursor")) mustBe BAD_REQUEST
    }

    "search by title or author, case-insensitively" in {
      ids("/api/v1/books?q=WELLS") mustBe Seq(10L, 6L)
      ids("/api/v1/books?q=sherlock") mustBe Seq(3L)
      ids("/api/v1/books?q=%25") mustBe empty // LIKE wildcards are matched literally
    }

    "filter by author, year and category, combined with AND" in {
      ids("/api/v1/books?author=h.%20g.%20wells") mustBe Seq(10L, 6L)
      ids("/api/v1/books?year=1818") mustBe Seq(4L)
      ids("/api/v1/books?category=horror") mustBe Seq(7L, 4L)
      ids("/api/v1/books?category=science-fiction&author=H.%20G.%20Wells") mustBe Seq(10L, 6L)
      ids("/api/v1/books?category=no-such-category") mustBe empty
    }

    "keep cursor pagination while filtering" in {
      val pages = walk("/api/v1/books", 2, "&category=children")
      pages.map(p => (p \ "data").as[Seq[JsObject]].size) mustBe Seq(2, 1)
      pages.flatMap(p => (p \ "data").as[Seq[JsObject]].map(b => (b \ "id").as[Long])) mustBe Seq(12L, 11L, 2L)
    }
  }

  "GET /api/v1/books/:id" should {

    "return book info with its chapter count" in {
      val book = getJson("/api/v1/books/1")
      (book \ "title").as[String] mustBe "Pride and Prejudice"
      (book \ "author").as[String] mustBe "Jane Austen"
      (book \ "chapterCount").as[Int] mustBe 5
      (book \ "categories").as[Seq[JsObject]].map(c => (c \ "slug").as[String]) mustBe Seq("classic", "romance")
      (book \ "coverUrl").toOption mustBe None
    }

    "return 404 for an unknown book" in {
      status(get("/api/v1/books/999")) mustBe NOT_FOUND
    }
  }

  "GET /api/v1/books/:id/chapters" should {

    "page through the content in reading order" in {
      val pages = walk("/api/v1/books/1/chapters", 2)
      pages.map(p => (p \ "data").as[Seq[JsObject]].size) mustBe Seq(2, 2, 1)
      val chapters = pages.flatMap(p => (p \ "data").as[Seq[JsObject]])
      chapters.map(c => (c \ "number").as[Int]) mustBe (1 to 5)
      (chapters.head \ "content").as[String] must startWith("It is a truth universally acknowledged")
    }

    "return an empty page for a book without content" in {
      val page = getJson("/api/v1/books/12/chapters")
      (page \ "data").as[Seq[JsObject]] mustBe empty
      (page \ "pagination" \ "hasMore").as[Boolean] mustBe false
    }

    "return 404 for an unknown book" in {
      status(get("/api/v1/books/999/chapters")) mustBe NOT_FOUND
    }

    "return 503 when a chapter's object is missing from storage" in {
      seeded
      val key  = "books/3/chapters/2.txt"
      val obj  = contentStore.objects.remove(key)
      try status(get("/api/v1/books/3/chapters")) mustBe SERVICE_UNAVAILABLE
      finally contentStore.objects.put(key, obj)
      status(get("/api/v1/books/3/chapters")) mustBe OK
    }
  }

  "GET /api/v1/books/:id/chapters/:number" should {

    "return a single chapter with its text from object storage" in {
      val chapter = getJson("/api/v1/books/5/chapters/1")
      (chapter \ "content").as[String] mustBe "Call me Ishmael."
      (chapter \ "sizeBytes").as[Long] mustBe 16L
      contentStore.text("books/5/chapters/1.txt") mustBe Some("Call me Ishmael.")
    }

    "return 404 for an unknown chapter" in {
      status(get("/api/v1/books/5/chapters/99")) mustBe NOT_FOUND
    }
  }

  "GET /api/v1/categories" should {

    "list categories by name" in {
      val slugs = (getJson("/api/v1/categories") \ "data").as[Seq[JsObject]].map(c => (c \ "slug").as[String])
      slugs must contain allOf ("classic", "horror", "science-fiction")
      slugs.size mustBe 9
    }

    "page through the books of one category" in {
      val pages = walk("/api/v1/categories/adventure/books", 2)
      pages.flatMap(p => (p \ "data").as[Seq[JsObject]].map(b => (b \ "id").as[Long])) mustBe Seq(12L, 11L, 5L)
    }

    "return 404 for an unknown category" in {
      status(get("/api/v1/categories/nope/books")) mustBe NOT_FOUND
    }
  }

  "GET /health/ready" should {

    "report ready when the database and storage answer" in {
      status(get("/health/ready")) mustBe OK
    }
  }

  "ContentSeeder" should {

    "create the sample catalogue and upload every chapter to storage" in {
      seeded mustBe 12
      contentStore.objects.size mustBe 15
    }
  }

  "Cursor" should {

    "round-trip and reject tampered values" in {
      Cursor.decode(Cursor.encode(42L)) mustBe Some(42L)
      Cursor.decode("") mustBe None
      Cursor.decode("%%%") mustBe None
      Cursor.decode(java.util.Base64.getUrlEncoder.encodeToString("v1:-1".getBytes)) mustBe None
    }
  }
}
