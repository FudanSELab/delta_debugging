package apiserver.request;

import java.util.ArrayList;

public class SetAsyncRequestSequenceRequest {
    private String clusterName;

    private ArrayList<String> svcList;

    public SetAsyncRequestSequenceRequest() {
        //do nothing
    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public SetAsyncRequestSequenceRequest(ArrayList<String> svcList) {
        this.svcList = svcList;
    }

    public ArrayList<String> getSvcList() {
        return svcList;
    }

    public void setSvcList(ArrayList<String> svcList) {
        this.svcList = svcList;
    }
}
