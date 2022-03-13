package registry

import (
	"demo/db"
	"demo/network/http"
	"demo/service/registry/model"
	"sort"
)

// svcDiscovery 服务发现
type svcDiscovery struct {
	db db.Db
}

func newSvcDiscovery(db db.Db) *svcDiscovery {
	return &svcDiscovery{db: db}
}

// 服务发现
func (s *svcDiscovery) discovery(req *http.Request) *http.Response {
	svcId, _ := req.QueryParam("service-id")
	svcType, _ := req.QueryParam("service-type")
	visitor := model.NewServiceProfileVisitor(svcId, model.ServiceType(svcType))
	result, err := s.db.QueryByVisitor(profileTable, visitor)
	if err != nil {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusInternalServerError).
			AddProblemDetails(err.Error())
	}
	profiles := make(profiles, 0)
	for _, record := range result {
		profile := record.(*model.ServiceProfileRecord).ToServiceProfile()
		region := new(model.Region)
		if err := s.db.Query(regionTable, profile.Region.Id, region); err != nil {
			return http.ResponseOfId(req.ReqId()).
				AddStatusCode(http.StatusInternalServerError).
				AddProblemDetails(err.Error())
		}
		profile.Region = region
		profiles.add(profile)
	}
	// 优先返回优先级高的，如果优先级相等，则返回负载较小的
	sort.Sort(profiles)
	if len(profiles) == 0 {
		return http.ResponseOfId(req.ReqId()).
			AddStatusCode(http.StatusNotFound)
	}
	return http.ResponseOfId(req.ReqId()).AddStatusCode(http.StatusOk).AddBody(profiles[0])
}

type profiles []*model.ServiceProfile

func (p profiles) Len() int {
	return len(p)
}

func (p profiles) Less(i, j int) bool {
	if p[i].Priority < p[j].Priority {
		return true
	} else if p[i].Priority == p[j].Priority {
		return p[i].Load < p[j].Load
	} else {
		return false
	}
}

func (p profiles) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *profiles) add(profile *model.ServiceProfile) {
	*p = append(*p, profile)
}
