from identified_service.building_blocks.db import get_mongo_database
from identified_service.gym_classes.service import GymClassService


class Config:
    @staticmethod
    def get_gym_class_service() -> GymClassService:
        return GymClassService(get_mongo_database())