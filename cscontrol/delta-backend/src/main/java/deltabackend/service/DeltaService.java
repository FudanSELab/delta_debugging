package deltabackend.service;

import deltabackend.domain.DeltaRequest;

public interface DeltaService {

    void delta(DeltaRequest message);
}
