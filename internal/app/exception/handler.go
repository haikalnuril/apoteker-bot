package exception

import (
	"errors"
	"telegram-doctor-recipe-helper-bot/internal/app/model"

	"github.com/gofiber/fiber/v2"
)

func Handler(ctx *fiber.Ctx, err error) error {
	var (
		badRequestError     *BadRequestError
		unauthorizedError   *UnauthorizedError
		forbiddenError      *ForbiddenError
		internalServerError *InternalServerError
		notFoundError       *NotFoundError
	)

	switch {
	case errors.As(err, &badRequestError):
		return ctx.Status(fiber.StatusBadRequest).JSON(model.Response{
			Code:    fiber.StatusBadRequest,
			Message: "Bad Request",
			Data:    badRequestError.Error(),
		})
	case errors.As(err, &unauthorizedError):
		return ctx.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Code:    fiber.StatusUnauthorized,
			Message: "Unauthorized",
			Data:    unauthorizedError.Error(),
		})
	case errors.As(err, &forbiddenError):
		return ctx.Status(fiber.StatusForbidden).JSON(model.Response{
			Code:    fiber.StatusForbidden,
			Message: "Forbidden",
			Data:    forbiddenError.Error(),
		})
	case errors.As(err, &notFoundError):
		return ctx.Status(fiber.StatusNotFound).JSON(model.Response{
			Code:    fiber.StatusNotFound,
			Message: "Not Found",
			Data:    notFoundError.Error(),
		})
	case errors.As(err, &internalServerError):
		return ctx.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Code:    fiber.StatusInternalServerError,
			Message: "Internal Server Error",
			Data:    internalServerError.Error(),
		})
	default:
		return ctx.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Code:    fiber.StatusInternalServerError,
			Message: "Internal Server Error",
			Data:    err.Error(),
		})
	}
}
