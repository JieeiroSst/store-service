package v1

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) initUploadRoutes(api fiber.Router) {
	group := api.Group("/upload")
	{
		group.Post("/", h.UploadFile)
		group.Put("/:id", h.Update)
		group.Get("/:id", h.GetByIdFile)
		group.Get("/", h.GetAll)
		group.Delete("/:id", h.Delete)
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

func (h *Handler) Update(ctx *fiber.Ctx) error {
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

	id := ctx.Params("id")

	if err := h.usecase.Update(ctx.Context(), id, fileMutipart, file); err != nil {
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

func (h *Handler) GetByIdFile(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	upload, err := h.usecase.GetById(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": false,
			"msg":   err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": true,
		"msg":   "FILE UPLOAD SUCCESS MESSAGE",
		"data":  upload,
	})
}

func (h *Handler) GetAll(ctx *fiber.Ctx) error {
	uploads, err := h.usecase.GetAll(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": false,
			"msg":   err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": true,
		"msg":   "FILE UPLOAD SUCCESS MESSAGE",
		"data":  uploads,
	})
}

func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := h.usecase.Delete(ctx.Context(), id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": false,
			"msg":   err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": true,
		"msg":   "DELETE FILE UPLOAD SUCCESS MESSAGE",
	})
}
