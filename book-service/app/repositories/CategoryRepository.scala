package repositories

import javax.inject._

import anorm.SqlParser._
import anorm._
import models.{Category, CategoryInput}
import play.api.db.Database

import scala.concurrent.Future

@Singleton
class CategoryRepository @Inject() (db: Database)(implicit ec: DatabaseExecutionContext) {

  private[repositories] val parser: RowParser[Category] =
    (long("id") ~ str("slug") ~ str("name")).map { case id ~ slug ~ name => Category(id, slug, name) }

  /** Categories are a short, curated list, so they are returned unpaginated. */
  def all(): Future[List[Category]] = Future {
    db.withConnection(implicit c => SQL"SELECT id, slug, name FROM categories ORDER BY name".as(parser.*))
  }

  def find(slug: String): Future[Option[Category]] = Future {
    db.withConnection(implicit c => SQL"SELECT id, slug, name FROM categories WHERE slug = $slug".as(parser.singleOpt))
  }

  def findBySlugs(slugs: Seq[String]): Future[List[Category]] =
    if (slugs.isEmpty) Future.successful(Nil)
    else
      Future {
        db.withConnection { implicit c =>
          SQL("SELECT id, slug, name FROM categories WHERE slug IN ({slugs})").on("slugs" -> slugs).as(parser.*)
        }
      }

  /** Left when the slug is taken. */
  def insert(input: CategoryInput): Future[Either[String, Category]] = Future {
    db.withConnection { implicit c =>
      try {
        val id = SQL"INSERT INTO categories (slug, name) VALUES (${input.slug}, ${input.name})"
          .executeInsert(get[Long](1).single)
        Right(Category(id, input.slug, input.name))
      } catch {
        case e: java.sql.SQLException if SqlUtil.isUniqueViolation(e) => Left(s"Category '${input.slug}' already exists")
      }
    }
  }

  def rename(slug: String, name: String): Future[Option[Category]] = Future {
    db.withConnection { implicit c =>
      if (SQL"UPDATE categories SET name = $name WHERE slug = $slug".executeUpdate() == 0) None
      else SQL"SELECT id, slug, name FROM categories WHERE slug = $slug".as(parser.singleOpt)
    }
  }

  /** Books keep existing; only their link to the category is removed. */
  def delete(slug: String): Future[Boolean] = Future {
    db.withConnection(implicit c => SQL"DELETE FROM categories WHERE slug = $slug".executeUpdate() > 0)
  }
}

private[repositories] object SqlUtil {

  /** SQLSTATE 23505 is unique_violation in both PostgreSQL and H2. */
  def isUniqueViolation(e: java.sql.SQLException): Boolean = e.getSQLState == "23505"

  /** Escapes LIKE wildcards; both PostgreSQL and H2 use backslash by default. */
  def likeContains(term: String): String =
    "%" + term.toLowerCase.replace("\\", "\\\\").replace("%", "\\%").replace("_", "\\_") + "%"
}
