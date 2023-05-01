from kink import di

from identified_service.building_blocks.clock import Clock
from identified_service.building_blocks.db import get_mongo_database
from identified_service.gym_passes.application.gym_pass_service import GymPassService
from identified_service.gym_passes.domain.gym_pass_repository import IGymPassRepository
from identified_service.gym_passes.facade import GymPassFacade
from identified_service.gym_passes.infrastructure.mongo_gym_pass_repository import MongoDBGymPassRepository


def bootstrap_di() -> None:
    repository = MongoDBGymPassRepository(get_mongo_database())
    di[IGymPassRepository] = repository
    di[GymPassService] = GymPassService(repository, Clock.system_clock())
    di[GymPassFacade] = GymPassFacade(repository)