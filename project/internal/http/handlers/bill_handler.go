package handlers

import (
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

type BillHandler struct {
	repo         repositories.BillRepository
	imageService image.Image
}

func NewBillHandler(repo repositories.BillRepository, imageService image.Image) *BillHandler {
	return &BillHandler{repo: repo, imageService: imageService}
}
