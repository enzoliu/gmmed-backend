package models

import (
	"breast-implant-warranty-system/internal/entity"

	"github.com/guregu/null/v5"
)

// Product search request
type ProductSearchRequest struct {
	ID            null.String `query:"id"`
	Brand         null.String `query:"brand"`
	Type          null.String `query:"type"`
	Active        null.Bool   `query:"active"`
	ModelNumber   null.String `query:"model_number"`
	Size          null.String `query:"size"`
	WarrantyYears null.Int    `query:"warranty_years"`
	entity.Pagination
}

type ProductMetadataAllResponse struct {
	Data []ProductMetadata `json:"data"`
}

type ProductMetadata struct {
	Brand       string `json:"brand"`
	Type        string `json:"type"`
	ModelNumber string `json:"model_number"`
	Size        string `json:"size"`
}
