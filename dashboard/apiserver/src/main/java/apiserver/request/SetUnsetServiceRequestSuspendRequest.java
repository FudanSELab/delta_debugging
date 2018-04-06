package apiserver.request;

public class SetUnsetServiceRequestSuspendRequest {

    public static final int SET_TO_SUSPEND = 1;

    public static final int UNSET_SUSPEND = 2;

    private String svc;

    private int actionType;

    public SetUnsetServiceRequestSuspendRequest() {
        //do nothing
    }

    public SetUnsetServiceRequestSuspendRequest(String svc, int actionType) {
        this.svc = svc;
        this.actionType = actionType;
    }

    public String getSvc() {
        return svc;
    }

    public void setSvc(String svc) {
        this.svc = svc;
    }

    public int getActionType() {
        return actionType;
    }

    public void setActionType(int actionType) {
        this.actionType = actionType;
    }
}
