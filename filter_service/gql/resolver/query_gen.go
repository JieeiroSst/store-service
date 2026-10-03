package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

func (r *queryResolver) Services(ctx context.Context) ([]*model.GatewayService, error) {
	out := make([]*model.GatewayService, 0, len(r.Clients.Services))
	for _, s := range r.Clients.Services {
		out = append(out, &model.GatewayService{Name: s.Name, Field: s.Field, EnvVar: s.EnvVar, BaseURL: s.BaseURL, Endpoints: s.Endpoints})
	}
	return out, nil
}

func (r *queryResolver) AccountTransactionService(ctx context.Context) (*model.AcctTxQuery, error) {
	return &model.AcctTxQuery{}, nil
}

func (r *queryResolver) AddressCountryService(ctx context.Context) (*model.AddrQuery, error) {
	return &model.AddrQuery{}, nil
}

func (r *queryResolver) AdmanagementService(ctx context.Context) (*model.AdmQuery, error) {
	return &model.AdmQuery{}, nil
}

func (r *queryResolver) AiAgentSystem(ctx context.Context) (*model.AiAgentQuery, error) {
	return &model.AiAgentQuery{}, nil
}

func (r *queryResolver) AirflowService(ctx context.Context) (*model.AirflowQuery, error) {
	return &model.AirflowQuery{}, nil
}

func (r *queryResolver) ArrangeService(ctx context.Context) (*model.ArrangeQuery, error) {
	return &model.ArrangeQuery{}, nil
}

func (r *queryResolver) AuthorizeService(ctx context.Context) (*model.AuthzQuery, error) {
	return &model.AuthzQuery{}, nil
}

func (r *queryResolver) AutomaticPaymentService(ctx context.Context) (*model.AutoPayQuery, error) {
	return &model.AutoPayQuery{}, nil
}

func (r *queryResolver) BankingService(ctx context.Context) (*model.BankQuery, error) {
	return &model.BankQuery{}, nil
}

func (r *queryResolver) BasketService(ctx context.Context) (*model.BasketQuery, error) {
	return &model.BasketQuery{}, nil
}

func (r *queryResolver) BillingService(ctx context.Context) (*model.BillingQuery, error) {
	return &model.BillingQuery{}, nil
}

func (r *queryResolver) BonuslinkService(ctx context.Context) (*model.BonusQuery, error) {
	return &model.BonusQuery{}, nil
}

func (r *queryResolver) BookService(ctx context.Context) (*model.BookSvcQuery, error) {
	return &model.BookSvcQuery{}, nil
}

func (r *queryResolver) BookStoreService(ctx context.Context) (*model.BookStoreQuery, error) {
	return &model.BookStoreQuery{}, nil
}

func (r *queryResolver) BookingMiniService(ctx context.Context) (*model.BookingMiniQuery, error) {
	return &model.BookingMiniQuery{}, nil
}

func (r *queryResolver) BotService(ctx context.Context) (*model.BotQuery, error) {
	return &model.BotQuery{}, nil
}

func (r *queryResolver) CalculateService(ctx context.Context) (*model.CalcQuery, error) {
	return &model.CalcQuery{}, nil
}

func (r *queryResolver) CallCenterAi(ctx context.Context) (*model.CallCenterQuery, error) {
	return &model.CallCenterQuery{}, nil
}

func (r *queryResolver) CarRentalService(ctx context.Context) (*model.CarRentalQuery, error) {
	return &model.CarRentalQuery{}, nil
}

func (r *queryResolver) CataloguesService(ctx context.Context) (*model.CatalogQuery, error) {
	return &model.CatalogQuery{}, nil
}

func (r *queryResolver) CdnService(ctx context.Context) (*model.CdnQuery, error) {
	return &model.CdnQuery{}, nil
}

func (r *queryResolver) ChatService(ctx context.Context) (*model.ChatQuery, error) {
	return &model.ChatQuery{}, nil
}

func (r *queryResolver) ChatbotSystem(ctx context.Context) (*model.ChatbotQuery, error) {
	return &model.ChatbotQuery{}, nil
}

func (r *queryResolver) CouponService(ctx context.Context) (*model.CouponQuery, error) {
	return &model.CouponQuery{}, nil
}

func (r *queryResolver) CustomerRelationshipService(ctx context.Context) (*model.CrmQuery, error) {
	return &model.CrmQuery{}, nil
}

func (r *queryResolver) DoordashService(ctx context.Context) (*model.DoordashQuery, error) {
	return &model.DoordashQuery{}, nil
}

func (r *queryResolver) DrawImageService(ctx context.Context) (*model.DrawImageQuery, error) {
	return &model.DrawImageQuery{}, nil
}

func (r *queryResolver) EkycService(ctx context.Context) (*model.EkycQuery, error) {
	return &model.EkycQuery{}, nil
}

func (r *queryResolver) FaceRecognitionService(ctx context.Context) (*model.FaceRecQuery, error) {
	return &model.FaceRecQuery{}, nil
}

func (r *queryResolver) FoodService(ctx context.Context) (*model.FoodQuery, error) {
	return &model.FoodQuery{}, nil
}

func (r *queryResolver) Geoservice(ctx context.Context) (*model.GeoQuery, error) {
	return &model.GeoQuery{}, nil
}

func (r *queryResolver) HospitalPatientManagementService(ctx context.Context) (*model.HospitalQuery, error) {
	return &model.HospitalQuery{}, nil
}

func (r *queryResolver) IdentifiedService(ctx context.Context) (*model.GymQuery, error) {
	return &model.GymQuery{}, nil
}

func (r *queryResolver) IntegratedPaymentService(ctx context.Context) (*model.IntPayQuery, error) {
	return &model.IntPayQuery{}, nil
}

func (r *queryResolver) KmsService(ctx context.Context) (*model.KmsQuery, error) {
	return &model.KmsQuery{}, nil
}

func (r *queryResolver) LivestreamService(ctx context.Context) (*model.LiveQuery, error) {
	return &model.LiveQuery{}, nil
}

func (r *queryResolver) LotteryService(ctx context.Context) (*model.LotteryQuery, error) {
	return &model.LotteryQuery{}, nil
}

func (r *queryResolver) ManageService(ctx context.Context) (*model.KcQuery, error) {
	return &model.KcQuery{}, nil
}

func (r *queryResolver) MediaService(ctx context.Context) (*model.MediaQuery, error) {
	return &model.MediaQuery{}, nil
}

func (r *queryResolver) MedicalService(ctx context.Context) (*model.MedicalQuery, error) {
	return &model.MedicalQuery{}, nil
}

func (r *queryResolver) MovieRecommendationService(ctx context.Context) (*model.MovieQuery, error) {
	return &model.MovieQuery{}, nil
}

func (r *queryResolver) NotificationService(ctx context.Context) (*model.NotifQuery, error) {
	return &model.NotifQuery{}, nil
}

func (r *queryResolver) NotifyhubService(ctx context.Context) (*model.HubQuery, error) {
	return &model.HubQuery{}, nil
}

func (r *queryResolver) OllamaModelPy(ctx context.Context) (*model.OllamaPyQuery, error) {
	return &model.OllamaPyQuery{}, nil
}

func (r *queryResolver) OllamaService(ctx context.Context) (*model.OllamaChatQuery, error) {
	return &model.OllamaChatQuery{}, nil
}

func (r *queryResolver) ParkingLotService(ctx context.Context) (*model.ParkingQuery, error) {
	return &model.ParkingQuery{}, nil
}

func (r *queryResolver) PartnerService(ctx context.Context) (*model.PartnerQuery, error) {
	return &model.PartnerQuery{}, nil
}

func (r *queryResolver) PayerService(ctx context.Context) (*model.PayerQuery, error) {
	return &model.PayerQuery{}, nil
}

func (r *queryResolver) PaymentWalletService(ctx context.Context) (*model.WalletQuery, error) {
	return &model.WalletQuery{}, nil
}

func (r *queryResolver) PaymentService(ctx context.Context) (*model.PaymentQuery, error) {
	return &model.PaymentQuery{}, nil
}

func (r *queryResolver) PhotoService(ctx context.Context) (*model.PhotoQuery, error) {
	return &model.PhotoQuery{}, nil
}

func (r *queryResolver) PointService(ctx context.Context) (*model.PointQuery, error) {
	return &model.PointQuery{}, nil
}

func (r *queryResolver) PolymarketService(ctx context.Context) (*model.PolyQuery, error) {
	return &model.PolyQuery{}, nil
}

func (r *queryResolver) PostService(ctx context.Context) (*model.PostQuery, error) {
	return &model.PostQuery{}, nil
}

func (r *queryResolver) QRService(ctx context.Context) (*model.QRQuery, error) {
	return &model.QRQuery{}, nil
}

func (r *queryResolver) RecompenseService(ctx context.Context) (*model.RecompQuery, error) {
	return &model.RecompQuery{}, nil
}

func (r *queryResolver) RecruitmentPlatformService(ctx context.Context) (*model.RecruitQuery, error) {
	return &model.RecruitQuery{}, nil
}

func (r *queryResolver) ReferralService(ctx context.Context) (*model.ReferralQuery, error) {
	return &model.ReferralQuery{}, nil
}

func (r *queryResolver) RentHouseService(ctx context.Context) (*model.RentQuery, error) {
	return &model.RentQuery{}, nil
}

func (r *queryResolver) ReservationService(ctx context.Context) (*model.ResvQuery, error) {
	return &model.ResvQuery{}, nil
}

func (r *queryResolver) RestaurantAccountingService(ctx context.Context) (*model.RAccountingQuery, error) {
	return &model.RAccountingQuery{}, nil
}

func (r *queryResolver) RestaurantConsumerService(ctx context.Context) (*model.RConsumerQuery, error) {
	return &model.RConsumerQuery{}, nil
}

func (r *queryResolver) RestaurantDeliveryService(ctx context.Context) (*model.RDeliveryQuery, error) {
	return &model.RDeliveryQuery{}, nil
}

func (r *queryResolver) RestaurantKitchenService(ctx context.Context) (*model.RKitchenQuery, error) {
	return &model.RKitchenQuery{}, nil
}

func (r *queryResolver) RestaurantOrderService(ctx context.Context) (*model.ROrderQuery, error) {
	return &model.ROrderQuery{}, nil
}

func (r *queryResolver) RewardService(ctx context.Context) (*model.RewardQuery, error) {
	return &model.RewardQuery{}, nil
}

func (r *queryResolver) RoomService(ctx context.Context) (*model.RoomQuery, error) {
	return &model.RoomQuery{}, nil
}

func (r *queryResolver) SardService(ctx context.Context) (*model.SardQuery, error) {
	return &model.SardQuery{}, nil
}

func (r *queryResolver) ShippingService(ctx context.Context) (*model.ShipQuery, error) {
	return &model.ShipQuery{}, nil
}

func (r *queryResolver) ShopifyService(ctx context.Context) (*model.ShopifyQuery, error) {
	return &model.ShopifyQuery{}, nil
}

func (r *queryResolver) ShortlinkService(ctx context.Context) (*model.ShortlinkQuery, error) {
	return &model.ShortlinkQuery{}, nil
}

func (r *queryResolver) SolanaService(ctx context.Context) (*model.SolQuery, error) {
	return &model.SolQuery{}, nil
}

func (r *queryResolver) ThreadsService(ctx context.Context) (*model.ThreadsQuery, error) {
	return &model.ThreadsQuery{}, nil
}

func (r *queryResolver) TicketService(ctx context.Context) (*model.TicketQuery, error) {
	return &model.TicketQuery{}, nil
}

func (r *queryResolver) ToggleService(ctx context.Context) (*model.ToggleQuery, error) {
	return &model.ToggleQuery{}, nil
}

func (r *queryResolver) ToolService(ctx context.Context) (*model.ToolQuery, error) {
	return &model.ToolQuery{}, nil
}

func (r *queryResolver) UploadService(ctx context.Context) (*model.UploadQuery, error) {
	return &model.UploadQuery{}, nil
}

func (r *queryResolver) UserService(ctx context.Context) (*model.UserQuery, error) {
	return &model.UserQuery{}, nil
}

func (r *queryResolver) VendingMachineService(ctx context.Context) (*model.VendQuery, error) {
	return &model.VendQuery{}, nil
}

func (r *queryResolver) VideoService(ctx context.Context) (*model.VideoQuery, error) {
	return &model.VideoQuery{}, nil
}

func (r *queryResolver) VoucherService(ctx context.Context) (*model.VoucherQuery, error) {
	return &model.VoucherQuery{}, nil
}

func (r *queryResolver) WalletService(ctx context.Context) (*model.WalletSvcQuery, error) {
	return &model.WalletSvcQuery{}, nil
}

func (r *queryResolver) WebrtcService(ctx context.Context) (*model.WebrtcQuery, error) {
	return &model.WebrtcQuery{}, nil
}

func (r *queryResolver) WishlistsService(ctx context.Context) (*model.WishlistQuery, error) {
	return &model.WishlistQuery{}, nil
}
