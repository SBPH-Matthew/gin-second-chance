package handlers

import (
	"net/http"
	"strconv"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/utils"
	"github.com/gin-gonic/gin"
)

func GetConversations(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var conversations []models.Conversation
	if err := database.DB.Preload("Product").
		Preload("Messages").
		Where("participant_one_id = ? OR participant_two_id = ?", userID, userID).
		Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch conversations"})
		return
	}

	baseURL := utils.GetBaseURL(c)
	for i := range conversations {
		if conversations[i].Product.ID != 0 {
			conversations[i].Product.Images = utils.FormatImageURLs(conversations[i].Product.Images, baseURL)
		}
	}

	c.JSON(http.StatusOK, conversations)
}

func GetMessages(c *gin.Context) {
	conversationID, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint)

	// Check if user is part of the conversation
	var conv models.Conversation
	if err := database.DB.Where("id = ? AND (participant_one_id = ? OR participant_two_id = ?)", conversationID, userID, userID).
		First(&conv).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}

	var messages []models.Message
	database.DB.Where("conversation_id = ?", conversationID).Order("created_at asc").Find(&messages)

	// Mark messages as read
	database.DB.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id != ?", conversationID, userID).
		Update("is_read", true)

	c.JSON(http.StatusOK, messages)
}

func SendMessage(c *gin.Context) {
	type MessageRequest struct {
		RecipientID uint   `json:"recipient_id" binding:"required"`
		ProductID   uint   `json:"product_id"`
		Content     string `json:"content" binding:"required"`
	}

	var body MessageRequest
	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	senderID := c.MustGet("userID").(uint)

	// Find or create conversation
	var conv models.Conversation
	err := database.DB.Where(
		"(participant_one_id = ? AND participant_two_id = ?) OR (participant_one_id = ? AND participant_two_id = ?)",
		senderID, body.RecipientID, body.RecipientID, senderID,
	).Where("product_id = ?", body.ProductID).First(&conv).Error

	if err != nil {
		conv = models.Conversation{
			ParticipantOneID: senderID,
			ParticipantTwoID: body.RecipientID,
			ProductID:        body.ProductID,
		}
		database.DB.Create(&conv)
	}

	message := models.Message{
		ConversationID: conv.ID,
		SenderID:       senderID,
		Content:        body.Content,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, message)
}

func GetAllConversations(c *gin.Context) {
	var conversations []models.Conversation
	if err := database.DB.Preload("Product").
		Preload("Messages").
		Order("updated_at desc").
		Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch conversations"})
		return
	}

	baseURL := utils.GetBaseURL(c)
	for i := range conversations {
		if conversations[i].Product.ID != 0 {
			conversations[i].Product.Images = utils.FormatImageURLs(conversations[i].Product.Images, baseURL)
		}
	}

	c.JSON(http.StatusOK, conversations)
}

func MakeOffer(c *gin.Context) {
	type OfferRequest struct {
		ConversationID uint    `json:"conversation_id" binding:"required"`
		OfferPrice     float64 `json:"offer_price" binding:"required"`
	}

	var body OfferRequest
	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	userID := c.MustGet("userID").(uint)

	var conv models.Conversation
	if err := database.DB.Preload("Product").Where("id = ? AND (participant_one_id = ? OR participant_two_id = ?)", body.ConversationID, userID, userID).
		First(&conv).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "Conversation not found or access denied"})
		return
	}

	// Update offer details
	conv.OfferPrice = body.OfferPrice
	conv.OfferStatus = "pending"

	if err := database.DB.Save(&conv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to make offer"})
		return
	}

	// Send a system message
	content := "I've made an offer of " + strconv.FormatFloat(body.OfferPrice, 'f', 2, 64)
	message := models.Message{
		ConversationID: conv.ID,
		SenderID:       userID,
		Content:        content,
	}
	database.DB.Create(&message)

	c.JSON(http.StatusOK, conv)
}

func AcceptOffer(c *gin.Context) {
	conversationID, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint)

	var conv models.Conversation
	if err := database.DB.Preload("Product").Where("id = ? AND (participant_one_id = ? OR participant_two_id = ?)", conversationID, userID, userID).
		First(&conv).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "Conversation not found or access denied"})
		return
	}

	// Only the seller can accept the offer
	// In SendMessage, participant_two is recipient. But conversation could be initiated by seller?
	// Usually participant_two is the one who was messaged.
	// Let's check sender of the offer. If sender was participant_one, then participant_two (seller) should accept.
	// Actually, the prompt says "Allow price negotiation while triggering a “Reserved” database state upon acceptance."

	if conv.Product.SellerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"message": "Only the seller can accept offers"})
		return
	}

	conv.OfferStatus = "accepted"
	if err := database.DB.Save(&conv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to accept offer"})
		return
	}

	// Trigger "Reserved" state in product
	var reservedStatus models.ProductStatus
	if err := database.DB.Where("name = ?", "RESERVED").First(&reservedStatus).Error; err == nil {
		database.DB.Model(&models.Product{}).Where("id = ?", conv.ProductID).Update("status_id", reservedStatus.ID)
	}

	// Send system message
	message := models.Message{
		ConversationID: conv.ID,
		SenderID:       userID,
		Content:        "Offer accepted! Product is now reserved for you.",
	}
	database.DB.Create(&message)

	c.JSON(http.StatusOK, conv)
}
