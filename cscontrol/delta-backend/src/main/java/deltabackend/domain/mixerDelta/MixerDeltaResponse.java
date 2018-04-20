package deltabackend.domain.mixerDelta;

import deltabackend.domain.api.request.SetAsyncRequestSequenceRequestWithSource;
import deltabackend.domain.bean.ServiceWithReplicas;
import deltabackend.domain.bean.SingleDeltaCMResourceRequest;
import deltabackend.domain.test.DeltaTestResponse;

import java.util.List;

public class MixerDeltaResponse {

    boolean status;
    String message;
    List<SingleDeltaCMResourceRequest> configEnv;
    List<SetAsyncRequestSequenceRequestWithSource> seqEnv;
    List<ServiceWithReplicas> instanceEnv;
    DeltaTestResponse result;
    boolean diffFromFirst;//different from the first test result, highlight it

    public MixerDeltaResponse(){

    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public List<SingleDeltaCMResourceRequest> getConfigEnv() {
        return configEnv;
    }

    public void setConfigEnv(List<SingleDeltaCMResourceRequest> configEnv) {
        this.configEnv = configEnv;
    }

    public List<SetAsyncRequestSequenceRequestWithSource> getSeqEnv() {
        return seqEnv;
    }

    public void setSeqEnv(List<SetAsyncRequestSequenceRequestWithSource> seqEnv) {
        this.seqEnv = seqEnv;
    }

    public List<ServiceWithReplicas> getInstanceEnv() {
        return instanceEnv;
    }

    public void setInstanceEnv(List<ServiceWithReplicas> instanceEnv) {
        this.instanceEnv = instanceEnv;
    }

    public DeltaTestResponse getResult() {
        return result;
    }

    public void setResult(DeltaTestResponse result) {
        this.result = result;
    }

    public boolean isDiffFromFirst() {
        return diffFromFirst;
    }

    public void setDiffFromFirst(boolean diffFromFirst) {
        this.diffFromFirst = diffFromFirst;
    }


}
