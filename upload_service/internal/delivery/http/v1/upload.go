package v1

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) initUploadRoutes(api fiber.Router) {
	group := api.Group("/upload")
	{
		group.Post("/", h.UploadFile)
	}

}

func (h *Handler) UploadFile(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("image")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": false,
			"msg":   err.Error(),
		})
	}

	fileMutipart, err := file.Open()
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": false,
			"msg":   err.Error(),
		})
	}

	receiverId := ctx.Query("receiver_id")

	err = h.usecase.Create(ctx.Context(), fileMutipart, file, receiverId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": false,
			"msg":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": true,
		"msg":   "FILE UPLOAD SUCCESS MESSAGE",
	})
}
