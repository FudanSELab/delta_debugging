package apiserver.service;

import apiserver.request.*;
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
    GetPodsListResponse getPodsList();
    GetPodsLogResponse getPodsLog();
    GetSinglePodLogResponse getSinglePodLog(GetSinglePodLogRequest getSinglePodLogRequest);
    RestartServiceResponse restartService();
}
