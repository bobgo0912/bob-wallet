package rpc

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/bobgo0912/b0b-common/pkg/log"
	"github.com/bobgo0912/b0b-common/pkg/server"
	"github.com/bobgo0912/bob-armory/pkg/wallet"
	"github.com/bobgo0912/bob-wallet/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WalletRpcServer struct {
	wallet.UnimplementedWalletServer
}

func RegService(s *server.GrpcServer) {
	s.RegService(&wallet.Wallet_ServiceDesc, &WalletRpcServer{})
}
func (w *WalletRpcServer) SettleHandle(ctx context.Context, req *wallet.SettleReq) (*wallet.SettleResp, error) {
	if len(req.Datas) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "Datas empty")
	}
	translations := make([]*model.Translation, 0, len(req.Datas))
	for _, data := range req.Datas {
		translation := model.Translation{
			Id:            data.CardId,
			Type:          1,
			Amount:        data.Amount,
			TranslationId: fmt.Sprintf("%d:%d", data.OrderId, data.CardId),
			Describe:      fmt.Sprintf("%d:%d prize", data.OrderId, data.CardId),
			PlayerId:      data.PlayerId,
			Status:        0,
		}
		translations = append(translations, &translation)
	}
	store, err := model.GetTranslationStore()
	if err != nil {
		log.Error("GetTranslationStore fail err=", err)
		return nil, err
	}
	err = store.MultipleInsert(ctx, translations)
	if err != nil {
		log.Error("MultipleInsert fail err=", err)
		return nil, err
	}
	return &wallet.SettleResp{Status: 0}, nil
}
func (w *WalletRpcServer) SettleCancel(ctx context.Context, req *wallet.SettleCancelReq) (*wallet.SettleCancelResp, error) {
	if len(req.Ids) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "Datas empty")
	}
	store, err := model.GetTranslationStore()
	if err != nil {
		log.Error("GetTranslationStore fail err=", err)
		return nil, err
	}
	list, err := store.QueryList(ctx, squirrel.Select("id", "player_id", "amount", "version").
		Where(squirrel.Eq{"id": req.Ids, "type": 1}))
	if err != nil {
		log.Error("QueryList fail err=", err)
		return nil, err
	}
	if len(list) < 1 {
		return nil, nil
	}
	tx, err := store.Tx(ctx)
	if err != nil {
		log.Error("Tx fail err=", err)
		return nil, err
	}
	m := make(map[uint64]uint64, 0)
	wids := make([]uint64, 0)
	for _, data := range list {
		wids = append(wids, data.PlayerId)
		amount, ok := m[data.PlayerId]
		if ok {
			m[data.PlayerId] = amount + data.Amount
		} else {
			m[data.PlayerId] = data.Amount
		}
		sql, param, err := squirrel.Update(store.TableName).SetMap(squirrel.Eq{"status": 2, "version": data.Version + 1}).
			Where(squirrel.Eq{"id": data.Id, "version": data.Version}).ToSql()
		if err != nil {
			log.Error("Update toSql err=", err)
			tx.Rollback()
			return nil, err
		}
		_, err = tx.Exec(sql, param...)
		if err != nil {
			log.Error("Update Exec err=", err)
			tx.Rollback()
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Error("Commit err=", err)
		return nil, err
	}
	return nil, nil
}
func (w *WalletRpcServer) SettleConfirm(ctx context.Context, req *wallet.SettleConfirmReq) (*wallet.SettleConfirmResp, error) {
	if len(req.Ids) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "Datas empty")
	}
	store, err := model.GetTranslationStore()
	if err != nil {
		log.Error("GetTranslationStore fail err=", err)
		return nil, err
	}
	list, err := store.QueryList(ctx, squirrel.Select("id", "player_id", "amount", "version").
		Where(squirrel.Eq{"id": req.Ids, "type": 1}))
	if err != nil {
		log.Error("QueryList fail err=", err)
		return nil, err
	}
	if len(list) < 1 {
		return nil, nil
	}
	tx, err := store.Tx(ctx)
	if err != nil {
		log.Error("Tx fail err=", err)
		return nil, err
	}
	m := make(map[uint64]uint64, 0)
	wids := make([]uint64, 0)
	for _, data := range list {
		wids = append(wids, data.PlayerId)
		amount, ok := m[data.PlayerId]
		if ok {
			m[data.PlayerId] = amount + data.Amount
		} else {
			m[data.PlayerId] = data.Amount
		}
		sql, param, err := squirrel.Update(store.TableName).SetMap(squirrel.Eq{"status": 1, "version": data.Version + 1}).
			Where(squirrel.Eq{"id": data.Id, "version": data.Version}).ToSql()
		if err != nil {
			log.Error("Update toSql err=", err)
			tx.Rollback()
			return nil, err
		}
		_, err = tx.Exec(sql, param...)
		if err != nil {
			log.Error("Update Exec err=", err)
			tx.Rollback()
			return nil, err
		}
	}
	walletStore, err := model.GetWalletStore()
	if err != nil {
		log.Error("GetWalletStore fail err=", err)
		tx.Rollback()
		return nil, err
	}
	queryList, err := walletStore.QueryList(ctx, squirrel.Select("id", "player_id", "balance", "version").Where(squirrel.Eq{"player_id": wids}))
	if err != nil {
		log.Error("QueryList fail err=", err)
		tx.Rollback()
		return nil, err
	}
	for _, w2 := range queryList {
		amount, ok := m[w2.PlayerId]
		if !ok {
			continue
		}
		a := w2.Balance + amount
		sql, param, err := squirrel.Update(model.WalletTableName).SetMap(squirrel.Eq{"balance": a, "version": w2.Version + 1}).
			Where(squirrel.Eq{"player_id": w2.PlayerId, "version": w2.Version}).ToSql()
		if err != nil {
			log.Error("Update toSql err=", err)
			tx.Rollback()
			return nil, err
		}
		_, err = tx.Exec(sql, param...)
		if err != nil {
			log.Error("Update Exec err=", err)
			tx.Rollback()
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Error("Commit err=", err)
		return nil, err
	}
	return nil, nil
}
