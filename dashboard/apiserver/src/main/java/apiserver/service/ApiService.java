package apiserver.service;

import apiserver.request.*;
import apiserver.response.*;
import org.springframework.web.bind.annotation.RequestBody;

public interface ApiService {
    GetClustersResponse getClusters();
    SetServiceReplicasResponse setServiceReplica(SetServiceReplicasRequest setServiceReplicasRequest);
    GetServicesListResponse getServicesList(String clusterName);
    GetServicesAndConfigResponse getServicesAndConfig(String clusterName);
    GetServiceReplicasResponse getServicesReplicas(GetServiceReplicasRequest getServiceReplicasRequest);
    SetRunOnSingleNodeResponse setRunOnSingleNode(String clusterName);
    ReserveServiceByListResponse reserveServiceByList(ReserveServiceRequest reserveServiceRequest);
    GetNodesListResponse getNodesList(String clusterName);
    DeltaNodeByListResponse deleteNodeByList(DeltaNodeRequest deltaNodeRequest);
    DeltaNodeByListResponse reserveNodeByList(DeltaNodeRequest deltaNodeRequest);
    GetPodsListResponse getPodsListAPI(String clusterName);
    GetPodsLogResponse getPodsLog(String clusterName);
    GetSinglePodLogResponse getSinglePodLog(GetSinglePodLogRequest getSinglePodLogRequest);
    RestartServiceResponse restartService(String clusterName);
    DeltaCMResourceResponse deltaCMResource(DeltaCMResourceRequest deltaCMResourceRequest);
    SetUnsetServiceRequestSuspendResponse setServiceRequestSuspend(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest);
    SetUnsetServiceRequestSuspendResponse unsetServiceRequestSuspend(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest);
    SetAsyncRequestSequenceResponse setAsyncRequestsSequence(SetAsyncRequestSequenceRequest setAsyncRequestSequenceRequest);
    SetUnsetServiceRequestSuspendResponse setServiceRequestSuspendWithSource(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest);
    ServiceWithEndpointsResponse getServiceWithEndpoints(String clusterName);
    ServiceWithEndpointsResponse getSpecificServiceWithEndpoints(ReserveServiceRequest reserveServiceRequest);
    SetAsyncRequestSequenceResponse setAsyncRequestsSequenceWithSource(SetAsyncRequestSequenceRequestWithSource setAsyncRequestSequenceRequest);
    SetAsyncRequestSequenceResponse setAsyncRequestSequenceWithSrcCombineWithFullSuspend(SetAsyncRequestSequenceRequestWithSource request);
}
