package service

import (
	"context"
	"fmt"
	"rates_service/infrastructure/prommetrics"
	"rates_service/internal/models"
	"rates_service/pkg/proto/gen/ratespb"
	servpb "rates_service/pkg/proto/gen/ratesservicepb"
	respb "rates_service/pkg/proto/gen/responsepb"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	provider interface {
		GetRates(ctx context.Context, market string) (models.RatesDTO, error)
	}
	repository interface {
		Create(ctx context.Context, rates models.RatesDTO) error
	}
)

type RateService struct {
	log    *zap.Logger
	market string
	remote provider
	repo   repository
	servpb.UnimplementedRatesServiceServer
}

func NewRateService(log *zap.Logger, market string, p provider, r repository) *RateService {
	return &RateService{
		log:    log,
		market: market,
		remote: p,
		repo:   r,
	}
}

func (rs *RateService) GetRates(ctx context.Context, req *servpb.GetRatesRequest) (*servpb.GetRatesResponse, error) {
	spanCtx, span := otel.GetTracerProvider().Tracer("rates_service").Start(ctx, "GetRates")
	defer span.End(trace.WithStackTrace(true))

	resp := &servpb.GetRatesResponse{
		ResponseMessage: &respb.ResponseMessage{
			Status: respb.STATUS_CODE_OK,
		}}

	mFuncMethod := prommetrics.ObcerveSummaryVecSplit("endpoints")
	defer func() {
		status := respb.STATUS_CODE_name[int32(resp.ResponseMessage.Status)]
		mFuncMethod("GetRates", status)
	}()

	mFuncProvider := prommetrics.ObcerveSummaryVecSplit("provider_API")
	rates, err := rs.remote.GetRates(spanCtx, rs.market)
	mFuncProvider("GetRates", func() string {
		if err != nil {
			return "success"
		}
		return "fail"
	}())
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

	mFuncDB := prommetrics.ObcerveSummaryVecSplit("DB")
	err = rs.repo.Create(spanCtx, rates)
	mFuncDB("rates", "Create", func() string {
		if err != nil {
			return "success"
		}
		return "fail"
	}())
	if err != nil {
		rs.log.Error(
			"RatesService",
			zap.String("method", "GetRates"),
			zap.NamedError("repository Create", err),
		)
		resp.ResponseMessage.Status = respb.STATUS_CODE_INTERNAL_ERROR
		resp.ResponseMessage.Message = fmt.Sprintf("repository call error: %s", err.Error())
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
