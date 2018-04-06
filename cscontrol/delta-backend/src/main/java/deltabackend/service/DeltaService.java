package deltabackend.service;

import deltabackend.domain.EnvParameter;
import deltabackend.domain.api.SetServiceReplicasRequest;
import deltabackend.domain.api.SetServiceReplicasResponse;
import deltabackend.domain.instanceDelta.DeltaRequest;
import deltabackend.domain.configDelta.ConfigDeltaRequest;
import deltabackend.domain.instanceDelta.SimpleInstanceRequest;
import deltabackend.domain.nodeDelta.DeltaNodeByListResponse;
import deltabackend.domain.nodeDelta.DeltaNodeRequest;
import deltabackend.domain.nodeDelta.NodeDeltaRequest;
import deltabackend.domain.sequenceDelta.SequenceDeltaRequest;
import deltabackend.domain.serviceDelta.ExtractServiceRequest;
import deltabackend.domain.serviceDelta.ReserveServiceResponse;
import deltabackend.domain.serviceDelta.ServiceDeltaRequest;

import java.util.List;

public interface DeltaService {

    void delta(DeltaRequest message);

    SetServiceReplicasResponse setInstanceNumOfServices(List<EnvParameter> env);

    void simpleSetInstance(SimpleInstanceRequest message);


    void serviceDelta(ServiceDeltaRequest message);

    void nodeDelta(NodeDeltaRequest message);

    void configDelta(ConfigDeltaRequest message);

    void sequenceDelta(SequenceDeltaRequest message);

    ReserveServiceResponse extractServices(ExtractServiceRequest testCases);

    DeltaNodeByListResponse deleteNodesByList(DeltaNodeRequest list);
}
