package service

import (
	"context"

	"galaxyotc/gc_services/exchangerate_service/exchangerates"

	pb "galaxyotc/common/proto/exchange_rate"
	"galaxyotc/common/log"
	"github.com/nats-io/go-nats"
)

type Service struct {
	fetcher *exchangerates.BitcoinPriceFetcher
}

// 创建新服务
func NewService() *Service {
	service := &Service{
		fetcher: exchangerates.NewBitcoinPriceFetcher(nil),
	}
	return service
}

// 服务启动
func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewExchangeRateServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}
	return nil
}

// 获取所有汇率
func (s *Service) GetAllRates(ctx context.Context, req pb.GetAllRatesReq) (pb.GetAllRatesResp, error) {
	log.Infof("Visiting GetAllRates, Request Params is %+v", req)
	allRates, err := s.fetcher.GetAllRates(req.Cache)
	if err != nil {
		log.Errorf("server-GetAllRates-Error: %s", err.Error())
		return pb.GetAllRatesResp{nil}, err
	}
	return pb.GetAllRatesResp{AllRates: allRates}, nil
}

// 获取指定币种的汇率（即，1个BTC的价格）
func (s *Service) GetExchangeRate(ctx context.Context, req pb.GetExchangeRateReq) (pb.GetExchangeRateResp, error) {
	log.Infof("Visiting GetExchangeRate, Request Params is %+v", req)
	rate, err := s.fetcher.GetExchangeRate(req.Code)
	if err != nil {
		log.Errorf("server-GetExchangeRate-Error: %s", err.Error())
		return pb.GetExchangeRateResp{}, err
	}
	return pb.GetExchangeRateResp{Rate: rate}, nil
}

// 获取最新指定币种的汇率（即，1个BTC的价格），并更新缓存
func (s *Service) GetLatestRate(ctx context.Context, req pb.GetExchangeRateReq) (pb.GetExchangeRateResp, error) {
	log.Infof("Visiting GetLatestRate, Request Params is %+v", req)
	rate, err := s.fetcher.GetLatestRate(req.Code)
	if err != nil {
		log.Errorf("server-GetLatestRate-Error: %s", err.Error())
		return pb.GetExchangeRateResp{}, err
	}
	return pb.GetExchangeRateResp{Rate: rate}, nil
}