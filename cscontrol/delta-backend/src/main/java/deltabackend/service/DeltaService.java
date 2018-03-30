package deltabackend.service;

import deltabackend.domain.DeltaRequest;
import deltabackend.domain.configDelta.ConfigDeltaRequest;
import deltabackend.domain.nodeDelta.DeltaNodeByListResponse;
import deltabackend.domain.nodeDelta.DeltaNodeRequest;
import deltabackend.domain.nodeDelta.NodeDeltaRequest;
import deltabackend.domain.sequenceDelta.SequenceDeltaRequest;
import deltabackend.domain.serviceDelta.ExtractServiceRequest;
import deltabackend.domain.serviceDelta.ReserveServiceByListResponse;
import deltabackend.domain.serviceDelta.ReserveServiceResponse;
import deltabackend.domain.serviceDelta.ServiceDeltaRequest;

import java.util.List;

public interface DeltaService {

    void delta(DeltaRequest message);

    void serviceDelta(ServiceDeltaRequest message);

    void nodeDelta(NodeDeltaRequest message);

    void configDelta(ConfigDeltaRequest message);

    void sequenceDelta(SequenceDeltaRequest message);

    ReserveServiceResponse extractServices(ExtractServiceRequest testCases);

    DeltaNodeByListResponse deleteNodesByList(DeltaNodeRequest list);
}
