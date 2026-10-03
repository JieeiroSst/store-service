package repositories

import javax.inject._

import akka.actor.ActorSystem
import play.api.libs.concurrent.CustomExecutionContext

/** Thread pool for blocking JDBC calls, configured as `database.dispatcher`. */
@Singleton
class DatabaseExecutionContext @Inject() (system: ActorSystem)
    extends CustomExecutionContext(system, "database.dispatcher")
