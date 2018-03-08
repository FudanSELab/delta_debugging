package apiserver.service;

import apiserver.request.DeltaNodeRequest;
import apiserver.request.GetServiceReplicasRequest;
import apiserver.request.ReserveServiceRequest;
import apiserver.request.SetServiceReplicasRequest;
import apiserver.response.*;

public interface ApiService {
    SetServiceReplicasResponse setServiceReplica(SetServiceReplicasRequest setServiceReplicasRequest);
    GetServicesListResponse getServicesList();
    GetServiceReplicasResponse getServicesReplicas(GetServiceReplicasRequest getServiceReplicasRequest);
    SetRunOnSingleNodeResponse setRunOnSingleNode();
    ReserveServiceByListResponse reserveServiceByList(ReserveServiceRequest reserveServiceRequest);
    GetNodesListResponse getNodesList();
    DeltaNodeByListResponse deleteNodeByList(DeltaNodeRequest deltaNodeRequest);
    DeltaNodeByListResponse reserveNodeByList(DeltaNodeRequest deltaNodeRequest);
}
