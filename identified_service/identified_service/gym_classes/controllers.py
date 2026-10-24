from bson.objectid import ObjectId
from fastapi import APIRouter, Depends, status
from fastapi.responses import JSONResponse

from identified_service.building_blocks.auth_deps import require_admin
from identified_service.gym_classes.config import Config
from identified_service.gym_classes.errors import GymClassNotFound
from identified_service.gym_classes.models import CreateOrUpdateClass, GymClass

router = APIRouter()


@router.post("/classes", response_model=GymClass, tags=["gym classes"], dependencies=[Depends(require_admin)])
async def create_class(gym_class_request: CreateOrUpdateClass) -> JSONResponse:
    service = Config.get_gym_class_service()
    result = service.create(gym_class_request)
    return JSONResponse(content=result.model_dump(), status_code=status.HTTP_201_CREATED)


@router.get("/classes", response_model=list[GymClass], tags=["gym classes"])
async def list_classes() -> JSONResponse:
    service = Config.get_gym_class_service()
    result = service.get_all()

    return JSONResponse(content=[item.model_dump() for item in result], status_code=status.HTTP_200_OK)


@router.get("/classes/{class_id}", response_model=GymClass, tags=["gym classes"])
async def get_class(class_id: str) -> JSONResponse:
    service = Config.get_gym_class_service()
    try:
        result = service.get(ObjectId(class_id))
    except GymClassNotFound as error:
        return JSONResponse(content={"message": str(error)}, status_code=status.HTTP_404_NOT_FOUND)

    return JSONResponse(content=result.model_dump(), status_code=status.HTTP_200_OK)


@router.put("/classes/{class_id}", response_model=GymClass, tags=["gym classes"], dependencies=[Depends(require_admin)])
async def update_class(class_id: str, gym_class_request: CreateOrUpdateClass) -> JSONResponse:
    service = Config.get_gym_class_service()
    try:
        result = service.update(class_id, gym_class_request)
    except GymClassNotFound as error:
        return JSONResponse(content={"message": str(error)}, status_code=status.HTTP_404_NOT_FOUND)

    return JSONResponse(content=result.model_dump(), status_code=status.HTTP_200_OK)


@router.delete("/classes/{class_id}", status_code=status.HTTP_204_NO_CONTENT, tags=["gym classes"], dependencies=[Depends(require_admin)])
async def delete_class(class_id: str) -> JSONResponse:
    service = Config.get_gym_class_service()
    service.delete(class_id)

    return JSONResponse(content={}, status_code=status.HTTP_204_NO_CONTENT)