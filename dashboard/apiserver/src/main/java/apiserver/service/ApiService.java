package apiserver.service;

import apiserver.request.SetServiceReplicasRequest;
import apiserver.response.SetServiceReplicasResponse;

public interface ApiService {
    SetServiceReplicasResponse setServiceReplica(SetServiceReplicasRequest setServiceReplicasRequest);
}
