import org.scalatestplus.play.PlaySpec
import org.scalatestplus.play.guice.GuiceOneAppPerSuite
import play.api.libs.json._
import play.api.mvc.Result
import play.api.test.FakeRequest
import play.api.test.Helpers._
import services.ContentSeeder

import scala.concurrent.duration._
import scala.concurrent.{Await, Future}

/** Write endpoints: books, chapters, covers and categories. */
class BookManagementSpec extends PlaySpec with GuiceOneAppPerSuite with support.TestApp {

  private lazy val seeded = Await.result(app.injector.instanceOf[ContentSeeder].done, 10.seconds)

  private def call(method: String, path: String, body: Option[JsValue] = None, key: Option[String] = Some(AdminKey))
      : Future[Result] = {
    seeded
    val base = FakeRequest(method, path).withHeaders(key.map("X-Api-Key" -> _).toSeq: _*)
    body match {
      case Some(json) => route(app, base.withJsonBody(json)).get
      case None       => route(app, base).get
    }
  }

  private def json(result: Future[Result]): JsValue = contentAsJson(result)

  private val dune = Json.obj(
    "title"         -> "  Dune  ",
    "author"        -> "Frank Herbert",
    "isbn"          -> "978-0-441-01359-3",
    "description"   -> "Desert planet politics.",
    "publishedYear" -> 1965,
    "categories"    -> Json.arr("science-fiction", "Classic")
  )

  "Write endpoints" should {

    "require the API key" in {
      status(call(POST, "/api/v1/books", Some(dune), key = None)) mustBe UNAUTHORIZED
      status(call(POST, "/api/v1/books", Some(dune), key = Some("wrong"))) mustBe UNAUTHORIZED
      status(call(DELETE, "/api/v1/books/1", key = None)) mustBe UNAUTHORIZED
    }
  }

  "Book CRUD" should {

    var duneId = 0L

    "create a book, normalising its fields" in {
      val result = call(POST, "/api/v1/books", Some(dune))
      status(result) mustBe CREATED
      val book = json(result)
      duneId = (book \ "id").as[Long]
      header(LOCATION, result) mustBe Some(s"/api/v1/books/$duneId")
      (book \ "title").as[String] mustBe "Dune"
      (book \ "isbn").as[String] mustBe "9780441013593"
      (book \ "categories").as[Seq[JsObject]].map(c => (c \ "slug").as[String]) mustBe Seq("classic", "science-fiction")
      (book \ "chapterCount").as[Int] mustBe 0
    }

    "list the new book first" in {
      val first = (json(call(GET, "/api/v1/books?limit=1")) \ "data")(0)
      (first \ "id").as[Long] mustBe duneId
    }

    "reject a duplicate ISBN with 409" in {
      status(call(POST, "/api/v1/books", Some(dune ++ Json.obj("title" -> "Dune again")))) mustBe CONFLICT
    }

    "reject invalid input with 400" in {
      status(call(POST, "/api/v1/books", Some(Json.obj("title" -> "", "author" -> "x")))) mustBe BAD_REQUEST
      status(call(POST, "/api/v1/books", Some(dune ++ Json.obj("isbn" -> "123")))) mustBe BAD_REQUEST
      status(call(POST, "/api/v1/books", Some(dune ++ Json.obj("isbn" -> "9780000000001", "publishedYear" -> -5)))) mustBe
        BAD_REQUEST
      val unknown = call(POST, "/api/v1/books", Some(dune ++ Json.obj("isbn" -> "9780000000002", "categories" -> Json.arr("nope"))))
      status(unknown) mustBe BAD_REQUEST
      (json(unknown) \ "error").as[String] must include("nope")
    }

    "replace a book's fields and categories" in {
      val updated = dune ++ Json.obj("title" -> "Dune (Deluxe)", "categories" -> Json.arr("adventure"))
      val result  = call(PUT, s"/api/v1/books/$duneId", Some(updated))
      status(result) mustBe OK
      (json(result) \ "title").as[String] mustBe "Dune (Deluxe)"
      (json(result) \ "categories").as[Seq[JsObject]].map(c => (c \ "slug").as[String]) mustBe Seq("adventure")
      status(call(PUT, "/api/v1/books/999", Some(updated))) mustBe NOT_FOUND
    }

    "write chapters to object storage" in {
      val c1 = call(PUT, s"/api/v1/books/$duneId/chapters/1", Some(Json.obj("title" -> "Book One", "content" -> "A beginning is a very delicate time.")))
      status(c1) mustBe CREATED
      contentStore.text(s"books/$duneId/chapters/1.txt") mustBe Some("A beginning is a very delicate time.")

      status(call(PUT, s"/api/v1/books/$duneId/chapters/2", Some(Json.obj("title" -> "Two", "content" -> "Fear is the mind-killer.")))) mustBe CREATED

      val replaced = call(PUT, s"/api/v1/books/$duneId/chapters/1", Some(Json.obj("title" -> "Book One", "content" -> "Ünïcödé text")))
      status(replaced) mustBe OK
      (json(replaced) \ "sizeBytes").as[Long] mustBe "Ünïcödé text".getBytes("UTF-8").length.toLong

      val page = json(call(GET, s"/api/v1/books/$duneId/chapters"))
      (page \ "data").as[Seq[JsObject]].map(c => (c \ "content").as[String]) mustBe Seq("Ünïcödé text", "Fear is the mind-killer.")
      (json(call(GET, s"/api/v1/books/$duneId")) \ "chapterCount").as[Int] mustBe 2
    }

    "reject bad chapters" in {
      status(call(PUT, s"/api/v1/books/$duneId/chapters/0", Some(Json.obj("title" -> "x", "content" -> "y")))) mustBe BAD_REQUEST
      status(call(PUT, s"/api/v1/books/$duneId/chapters/3", Some(Json.obj("title" -> "x", "content" -> "  ")))) mustBe BAD_REQUEST
      status(call(PUT, "/api/v1/books/999/chapters/1", Some(Json.obj("title" -> "x", "content" -> "y")))) mustBe NOT_FOUND
    }

    "delete a chapter and its object" in {
      status(call(DELETE, s"/api/v1/books/$duneId/chapters/2")) mustBe NO_CONTENT
      contentStore.objects.containsKey(s"books/$duneId/chapters/2.txt") mustBe false
      status(call(GET, s"/api/v1/books/$duneId/chapters/2")) mustBe NOT_FOUND
      status(call(DELETE, s"/api/v1/books/$duneId/chapters/2")) mustBe NOT_FOUND
    }

    "store a cover in object storage and serve it" in {
      val png = Array[Byte](0x89.toByte, 0x50, 0x4e, 0x47, 1, 2, 3)
      def upload(bytes: Array[Byte], contentType: String) = {
        seeded
        route(
          app,
          FakeRequest(PUT, s"/api/v1/books/$duneId/cover")
            .withHeaders("X-Api-Key" -> AdminKey, CONTENT_TYPE -> contentType)
            .withBody(akka.util.ByteString(bytes))
        ).get
      }

      status(upload(png, "text/plain")) mustBe BAD_REQUEST

      val result = upload(png, "image/png")
      status(result) mustBe OK
      // No public MinIO endpoint in tests, so the URL is the API's proxy.
      (json(result) \ "coverUrl").as[String] mustBe s"/api/v1/books/$duneId/cover"

      val served = call(GET, s"/api/v1/books/$duneId/cover")
      status(served) mustBe OK
      contentType(served) mustBe Some("image/png")
      contentAsBytes(served).toArray mustBe png

      // Replacing the cover removes the previous object.
      status(upload(Array[Byte](1, 2), "image/jpeg")) mustBe OK
      contentStore.objects.keySet.toArray.count(_.toString.startsWith(s"books/$duneId/cover")) mustBe 1

      status(call(DELETE, s"/api/v1/books/$duneId/cover")) mustBe NO_CONTENT
      status(call(GET, s"/api/v1/books/$duneId/cover")) mustBe NOT_FOUND
      (json(call(GET, s"/api/v1/books/$duneId")) \ "coverUrl").toOption mustBe None
    }

    "delete a book with its objects" in {
      status(call(DELETE, s"/api/v1/books/$duneId")) mustBe NO_CONTENT
      status(call(GET, s"/api/v1/books/$duneId")) mustBe NOT_FOUND
      contentStore.objects.keySet.toArray.exists(_.toString.startsWith(s"books/$duneId/")) mustBe false
      status(call(DELETE, s"/api/v1/books/$duneId")) mustBe NOT_FOUND
    }
  }

  "Category management" should {

    "create, rename and delete a category" in {
      val created = call(POST, "/api/v1/categories", Some(Json.obj("slug" -> "Poetry", "name" -> "Poetry")))
      status(created) mustBe CREATED
      (json(created) \ "slug").as[String] mustBe "poetry"
      status(call(POST, "/api/v1/categories", Some(Json.obj("slug" -> "poetry", "name" -> "Again")))) mustBe CONFLICT
      status(call(POST, "/api/v1/categories", Some(Json.obj("slug" -> "bad slug!", "name" -> "x")))) mustBe BAD_REQUEST

      (json(call(PUT, "/api/v1/categories/poetry", Some(Json.obj("name" -> "Poems")))) \ "name").as[String] mustBe "Poems"

      status(call(DELETE, "/api/v1/categories/poetry")) mustBe NO_CONTENT
      status(call(GET, "/api/v1/categories/poetry")) mustBe NOT_FOUND
    }

    "keep books when their category is deleted" in {
      status(call(DELETE, "/api/v1/categories/horror")) mustBe NO_CONTENT
      val dracula = json(call(GET, "/api/v1/books/7"))
      (dracula \ "categories").as[Seq[JsObject]].map(c => (c \ "slug").as[String]) mustBe Seq("classic")
    }
  }
}
