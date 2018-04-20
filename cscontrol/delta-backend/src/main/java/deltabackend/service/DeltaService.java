package deltabackend.service;

import deltabackend.domain.api.request.DeltaNodeRequest;
import deltabackend.domain.api.response.DeltaNodeByListResponse;
import deltabackend.domain.api.response.SetServiceReplicasResponse;
import deltabackend.domain.bean.ServiceReplicasSetting;
import deltabackend.domain.configDelta.ConfigDeltaRequest;
import deltabackend.domain.instanceDelta.DeltaRequest;
import deltabackend.domain.instanceDelta.SimpleInstanceRequest;
import deltabackend.domain.mixerDelta.MixerDeltaRequest;
import deltabackend.domain.nodeDelta.NodeDeltaRequest;
import deltabackend.domain.sequenceDelta.SequenceDeltaRequest;
import deltabackend.domain.serviceDelta.ExtractServiceRequest;
import deltabackend.domain.serviceDelta.ReserveServiceResponse;
import deltabackend.domain.serviceDelta.ServiceDeltaRequest;

import java.util.List;
import java.util.concurrent.ExecutionException;

public interface DeltaService {

    void delta(DeltaRequest message) throws ExecutionException, InterruptedException;

    SetServiceReplicasResponse setInstanceNumOfServices(List<ServiceReplicasSetting> env, String cluster);

    void simpleSetInstance(SimpleInstanceRequest message);


    void serviceDelta(ServiceDeltaRequest message);

    void nodeDelta(NodeDeltaRequest message);

    void configDelta(ConfigDeltaRequest message) throws ExecutionException, InterruptedException;

    void simpleSetOrignal(ConfigDeltaRequest message);

    void sequenceDelta(SequenceDeltaRequest message) throws ExecutionException, InterruptedException;

    void mixerDelta(MixerDeltaRequest message) throws ExecutionException, InterruptedException;

//    ReserveServiceResponse extractServices(ExtractServiceRequest testCases);

//    DeltaNodeByListResponse deleteNodesByList(DeltaNodeRequest list);
}
