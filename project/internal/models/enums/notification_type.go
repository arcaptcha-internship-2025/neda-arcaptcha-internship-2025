package enums

type NotificationType string

const (
	BillNotification         NotificationType = "bill"
	AnnouncementNotification NotificationType = "announcement"
	PaymentNotification      NotificationType = "payment"
)
