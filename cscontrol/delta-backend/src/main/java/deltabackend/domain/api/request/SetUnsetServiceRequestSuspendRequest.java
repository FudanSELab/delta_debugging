package deltabackend.domain.api.request;

public class SetUnsetServiceRequestSuspendRequest {

    public static final int SET_TO_SUSPEND = 1;

    public static final int UNSET_SUSPEND = 2;

    private String clusterName;

    private String sourceSvcName;

    private String svc;


    public SetUnsetServiceRequestSuspendRequest() {
        //do nothing
    }

    public SetUnsetServiceRequestSuspendRequest(String clusterName, String sourceSvcName, String svc) {
        this.clusterName = clusterName;
        this.sourceSvcName = sourceSvcName;
        this.svc = svc;
    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public String getSourceSvcName() {
        return sourceSvcName;
    }

    public void setSourceSvcName(String sourceSvcName) {
        this.sourceSvcName = sourceSvcName;
    }

    public String getSvc() {
        return svc;
    }

    public void setSvc(String svc) {
        this.svc = svc;
    }

}
