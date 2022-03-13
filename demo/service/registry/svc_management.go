package registry

import (
	"demo/db"
	"demo/network/http"
	"demo/service/registry/model"
	"demo/sidecar"
	"fmt"
	"github.com/google/uuid"
)

// svcManagement 服务管理，包含服务注册、更新、去注册。另外，服务订阅、去订阅、通知的功能由于与服务注册、更新、去注册紧密关联，
// 比如，每次的服务通知都是发生在服务状态变更之后，因此也把它们归到服务管理模块。
type svcManagement struct {
	localIp        string
	db             db.Db
	sidecarFactory sidecar.Factory
}

func newSvcManagement(localIp string, db db.Db, sidecarFactory sidecar.Factory) *svcManagement {
	return &svcManagement{
		localIp:        localIp,
		db:             db,
		sidecarFactory: sidecarFactory,
	}
}

// 服务注册
func (s *svcManagement) register(req *http.Request) *http.Response {
	profile, ok := req.Body().(*model.ServiceProfile)
	if !ok {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusBadRequest).
			AddProblemDetails("service register request's body is not *ServiceProfile")
	}
	transaction := s.db.CreateTransaction("register" + profile.Id)
	transaction.Begin()
	region := new(model.Region)
	// 因为Region表是被关联的，如果Region不存在了，就插入一条记录
	if err := s.db.Query(regionTable, profile.Region.Id, region); err != nil {
		cmd := db.NewInsertCmd(regionTable).WithPrimaryKey(profile.Region.Id).WithRecord(profile.Region)
		transaction.Exec(cmd)
	}
	cmd := db.NewInsertCmd(profileTable).WithPrimaryKey(profile.Id).WithRecord(profile.ToTableRecord())
	transaction.Exec(cmd)

	if err := transaction.Commit(); err != nil {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	// 发送通知
	go s.notify(model.Register, profile)
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusCreate)
}

// 服务更新
func (s *svcManagement) update(req *http.Request) *http.Response {
	profile, ok := req.Body().(*model.ServiceProfile)
	if !ok {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusBadRequest).
			AddProblemDetails("service update request's body is not *ServiceProfile")
	}
	transaction := s.db.CreateTransaction("register" + profile.Id)
	transaction.Begin()
	// 先更新regions表
	rcmd := db.NewUpdateCmd(regionTable).WithPrimaryKey(profile.Region.Id).WithRecord(profile.Region)
	transaction.Exec(rcmd)
	pcmd := db.NewUpdateCmd(profileTable).WithPrimaryKey(profile.Id).WithRecord(profile.ToTableRecord())
	transaction.Exec(pcmd)
	if err := transaction.Commit(); err != nil {
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	// 发送通知
	go s.notify(model.Update, profile)
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusOk)
}

// 去注册
func (s *svcManagement) deregister(req *http.Request) *http.Response {
	svcId, ok := req.Header("service-id")
	if !ok {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusBadRequest).
			AddProblemDetails("service deregister request not contain service-id header")
	}
	profileRecord := new(model.ServiceProfileRecord)
	var profile *model.ServiceProfile
	if err := s.db.Query(profileTable, svcId, profileRecord); err == nil {
		profile = profileRecord.ToServiceProfile()
		region := new(model.Region)
		if err = s.db.Query(regionTable, profile.Region.Id, region); err != nil {
			return http.ResponseOfId(req.ReqId()).
				AddStatusCode(http.StatusInternalServerError).AddProblemDetails(err.Error())
		}
		profile.Region = region
		if err := s.db.Delete(profileTable, svcId); err != nil {
			return http.ResponseOfId(req.ReqId()).
				AddStatusCode(http.StatusInternalServerError).AddProblemDetails(err.Error())
		}
		// 发送通知
		go s.notify(model.Deregister, profile)
		return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusNoContent)
	}
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusBadRequest).
		AddProblemDetails("service-id " + svcId + " not exist")
}

// 服务订阅
func (s *svcManagement) subscribe(req *http.Request) *http.Response {
	subscription, ok := req.Body().(*model.Subscription)
	if !ok {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusBadRequest).
			AddProblemDetails("subscribe request's body is not Subscription")
	}
	if subscription.Id != "" {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusBadRequest).
			AddProblemDetails("Subscription Id is not empty")
	}
	subscription.Id = uuid.NewString()
	if err := s.db.Insert(subscriptionTable, subscription.Id, subscription); err != nil {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusCreate).
		AddHeader("subscription-id", subscription.Id)
}

// 服务去订阅
func (s *svcManagement) unsubscribe(req *http.Request) *http.Response {
	subscriptionId, ok := req.Header("subscription-id")
	if !ok {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusBadRequest).
			AddProblemDetails("service unsubscribe request not contain subscription-id header")
	}
	if err := s.db.Delete(subscriptionTable, subscriptionId); err != nil {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusInternalServerError).AddProblemDetails(err.Error())
	}
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusNoContent)
}

// 服务通知
func (s *svcManagement) notify(notifyType model.NotifyType, profile *model.ServiceProfile) {
	visitor := model.NewSubscriptionVisitor(profile.Id, profile.Type)
	result, err := s.db.QueryByVisitor(subscriptionTable, visitor)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	httpClient, err := http.NewClient(s.sidecarFactory.Create(), s.localIp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, record := range result {
		subscription := record.(*model.Subscription)
		notification := model.NewNotification(subscription.Id)
		notification.Type = notifyType
		notification.Profile = profile.Clone().(*model.ServiceProfile)
		notifyUri, err := subscription.NotifyUri()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		notifyEndpoint, err := subscription.NotifyEndpoint()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		req := http.EmptyRequest().AddUri(http.Uri(notifyUri)).AddMethod(http.POST).AddBody(notification)
		resp, err := httpClient.Send(notifyEndpoint, req)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("notify %s success, resp %+v", subscription.SrcSvcId, resp)
	}
}
