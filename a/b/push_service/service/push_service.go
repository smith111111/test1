package service

import (
	"fmt"
	"time"
	"encoding/json"

	"galaxyotc/common/log"
	"galaxyotc/common/utils"
	pb "galaxyotc/common/proto/push"

	"github.com/panjf2000/ants"
	"galaxyotc/common/model"
	"galaxyotc/gc_services/push_service/push"
)

type PushService struct {
	TS	*TokenService
	AS 	*AppkeyService
	AP 	*ants.Pool
}

func newPushService() *PushService {
	ap, _ := ants.NewPool(200)
	service := &PushService{
		TS: newTokenService(),
		AS: newAppkeyService(),
		AP: ap,
	}
	return service
}

//列播(含单播)
func (s *PushService) Push(req *pb.SendMsgReq) (bool, string) {
	//map[推送系统类型]map[设备类型]map[推送AppKey&推送Secret]设备唯一标识
	pushSysMap := make(map[int32]map[int32]map[string]*model.DeviceTokenAndPackageName)

	receivers := utils.SplitToUint(req.Receivers, ",")

	for _, receiver := range receivers {
		if receiver == 0 {
			continue
		}

		pushTokenItem := s.TS.GetTokenCache(fmt.Sprintf("%d-%d", req.AppId, receiver))
		if pushTokenItem == nil {
			log.Warnf("要推送的uid其token不存在, uid :%d", receiver)
			continue
		}

		log.Debugf("PushTokenItem:%v\n", pushTokenItem)

		//登录状态满足
		if pushTokenItem.LoginStatus == req.LoginStatus {
			//将pushAppKey作为map的key,pushToken组成字符串为value
			pushAppKeyItem := s.AS.GetAppkeyItem(req.AppId, pushTokenItem.Platform, pushTokenItem.PushType)
			if pushAppKeyItem == nil {
				log.Warnf("要推送的app平台其appkey不存在, appid:%d, platform:%d, pushSysType:%d", req.AppId, pushTokenItem.Platform, pushTokenItem.PushType)
				continue
			}

			log.Debugf("PushAppKeyItem:%v\n", pushAppKeyItem)

			platMap, ok := pushSysMap[pushTokenItem.PushType]
			if !ok {
				pushSysMap[pushTokenItem.PushType] = make(map[int32]map[string]*model.DeviceTokenAndPackageName)
				platMap, _ = pushSysMap[pushTokenItem.PushType]
			}

			appKeySecret, ok := platMap[pushTokenItem.Platform]
			if !ok {
				platMap[pushTokenItem.Platform] = make(map[string]*model.DeviceTokenAndPackageName)
				appKeySecret, _ = platMap[pushTokenItem.Platform]
			}

			key := fmt.Sprintf("%s&%s", pushAppKeyItem.Appkey, pushAppKeyItem.Secret)
			_, ok = appKeySecret[key]
			if ok {
				item := appKeySecret[key]
				item.DeviceTokens += "," + pushTokenItem.DeviceToken
				item.UserIds = append(item.UserIds, receiver)
				item.MsgIds = append(item.MsgIds, model.GetPushMsgId())
				item.PackageName = pushAppKeyItem.PackageName
			} else {
				item := &model.DeviceTokenAndPackageName{}
				item.DeviceTokens = pushTokenItem.DeviceToken
				item.UserIds = make([]uint, 0)
				item.UserIds = append(item.UserIds, receiver)
				item.MsgIds = make([]int64, 0)
				item.MsgIds = append(item.MsgIds, model.GetPushMsgId())
				item.PackageName = pushAppKeyItem.PackageName
				appKeySecret[key] = item
			}
		} else {
			log.Warnf("appid:%d, uid: %d, 未登录", req.AppId, receiver)
		}
	}

	if len(pushSysMap) == 0 {
		return false, "推送失败"
	}

	s.push(req, pushSysMap)

	return true, "success"
}

func (s *PushService) push(req *pb.SendMsgReq, pushSysMap map[int32]map[int32]map[string]*model.DeviceTokenAndPackageName) {

	//推送要用到的数据
	info := &model.PushInfo{
		AppId: req.AppId,
		Title: req.Title,
		Text: req.Text,
	}

	custom := make(map[string]interface{})

	err := json.Unmarshal(req.Custom, &custom)
	if err != nil {
		log.Errorf("PushService-push-Error: %s", err.Error())
	}

	info.Custom = custom

	info.DisplayType = req.DisplayType
	info.ExpireTime = time.Unix(int64(req.ExpireTime), 0).Format("2006-01-02 15:04:05")

	//存到数据库的消息记录
	pushMsg := &model.PushMsg{
		AppId: req.AppId,
		Title: req.Title,
		Text: req.Text,
		Custom: string(req.Custom),
		LoginStatus: req.LoginStatus,
		PushMode: model.EPUSH_MSG_TYPE_LISTCAST,
	}

	for pushType, platMap := range pushSysMap {
		for platform, appKeySecrets := range platMap {
			info.AppKeySecrets = appKeySecrets

			s.AP.Submit(func() error {
				service := push.GetPushService(model.EPUSH_SYS_TYPE(pushType))
				service.Push(info, model.EPUSH_MSG_TYPE_LISTCAST, model.EDEVICE_TYPE(platform), func(changeStatus, pushId, errorCode string) {
					log.Debugf("changeStatus:%s, pushId:%s, errorCode:%s\n", changeStatus, pushId, errorCode)

					//更新状态
					for _, v := range info.AppKeySecrets {
						for _, msgId := range v.MsgIds {
							updateTime := time.Now().Unix()
							if err := model.DB.Table("push_msg").Where(model.PushMsg{MsgId: msgId}).Updates(&model.PushMsg{PushStatus: changeStatus, ErrorCode: errorCode, PushId: pushId, UpdateTime: int32(updateTime)}).Error; err != nil {
								log.Errorf("PushService-push-Error: %s", err.Error())
								continue
							}
						}
					}
				})

				return nil
			})

			//保存记录
			for _, v := range appKeySecrets {
				for i, userId := range v.UserIds {
					pushMsg.Platform = platform
					pushMsg.MsgId = v.MsgIds[i]
					pushMsg.UserId = userId

					if err := model.DB.Create(&pushMsg).Error; err != nil {
						log.Errorf("PushService-push-Error: %s", err.Error())
						continue
					}

				}
			}
		}
	}
}
