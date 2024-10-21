package service

import (
	"context"
	"fmt"
	"rates_service/internal/models"
	"rates_service/pkg/proto/gen/ratespb"
	servpb "rates_service/pkg/proto/gen/ratesservicepb"
	respb "rates_service/pkg/proto/gen/responsepb"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type provider interface {
	GetRates(ctx context.Context, market string) (models.RatesDTO, error)
}

type RateService struct {
	market string
	remote provider
	log    *zap.Logger
	db     *sqlx.DB
	servpb.UnimplementedRatesServiceServer
}

func NewRateService(market string, p provider, log *zap.Logger, db *sqlx.DB) *RateService {
	return &RateService{
		market: market,
		remote: p,
		log:    log,
		db:     db,
	}
}

func (rs *RateService) GetRates(ctx context.Context, req *servpb.GetRatesRequest) (*servpb.GetRatesResponse, error) {
	resp := &servpb.GetRatesResponse{
		ResponseMessage: &respb.ResponseMessage{
			Status: respb.STATUS_CODE_OK,
		}}
	rates, err := rs.remote.GetRates(ctx, rs.market)
	if err != nil {
		rs.log.Error(
			"RatesService",
			zap.String("method", "GetRates"),
			zap.NamedError("provider GetRates", err),
		)
		resp.ResponseMessage.Status = respb.STATUS_CODE_INTERNAL_ERROR
		resp.ResponseMessage.Message = fmt.Sprintf("remote service call error: %s", err.Error())
		return resp, nil
	}
	resp.Rates = &ratespb.Rates{
		Timestamp: &timestamppb.Timestamp{
			Seconds: rates.Timestamp,
		},
		AskPrice: rates.AskPrice,
		BidPrice: rates.BidPrice,
	}
	return resp, nil
}
