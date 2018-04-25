package apiserver.service;

import apiserver.request.*;
import apiserver.response.*;

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
    SetUnsetServiceRequestSuspendResponse unsetServiceRequestSuspendWithSource(SetUnsetServiceRequestSuspendRequest request);
    SetAsyncRequestSequenceResponse setAsyncRequestsSequence(SetAsyncRequestSequenceRequest setAsyncRequestSequenceRequest);
    SetUnsetServiceRequestSuspendResponse setServiceRequestSuspendWithSource(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest);
    ServiceWithEndpointsResponse getServiceWithEndpoints(String clusterName);
    ServiceWithEndpointsResponse getSpecificServiceWithEndpoints(ReserveServiceRequest reserveServiceRequest);
    SetAsyncRequestSequenceResponse setAsyncRequestsSequenceWithSource(SetAsyncRequestSequenceRequestWithSource setAsyncRequestSequenceRequest);
    SetAsyncRequestSequenceResponse setAsyncRequestSequenceWithSrcCombineWithFullSuspend(SetAsyncRequestSequenceRequestWithSource request);
    SetAsyncRequestSequenceResponse unsuspendAllRequest(SetAsyncRequestSequenceRequestWithSource request);
    SimpleResponse deltaAll(DeltaAllRequest deltaAllRequest);
}
