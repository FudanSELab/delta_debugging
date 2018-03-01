package apiserver.service;

import apiserver.request.GetServiceReplicasRequest;
import apiserver.request.SetServiceReplicasRequest;
import apiserver.response.GetServiceReplicasResponse;
import apiserver.response.GetServicesListResponse;
import apiserver.response.SetRunOnSingleNodeResponse;
import apiserver.response.SetServiceReplicasResponse;

public interface ApiService {
    SetServiceReplicasResponse setServiceReplica(SetServiceReplicasRequest setServiceReplicasRequest);
    GetServicesListResponse getServicesList();
    GetServiceReplicasResponse getServicesReplicas(GetServiceReplicasRequest getServiceReplicasRequest);
    SetRunOnSingleNodeResponse setRunOnSingleNode();
}
