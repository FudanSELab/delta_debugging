package apiserver.service;

import apiserver.request.*;
import apiserver.response.*;

public interface ApiService {
    SetServiceReplicasResponse setServiceReplica(SetServiceReplicasRequest setServiceReplicasRequest);
    GetServicesListResponse getServicesList();
    GetServicesAndConfigResponse getServicesAndConfig();
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
    DeltaCMResourceResponse deltaCMResource(DeltaCMResourceRequest deltaCMResourceRequest);
    SetUnsetServiceRequestSuspendResponse setServiceRequestSuspend(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest);
    SetUnsetServiceRequestSuspendResponse unsetServiceRequestSuspend(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest);
    SetAsyncRequestSequenceResponse setAsyncRequestsSequence(SetAsyncRequestSequenceRequest setAsyncRequestSequenceRequest);
}
