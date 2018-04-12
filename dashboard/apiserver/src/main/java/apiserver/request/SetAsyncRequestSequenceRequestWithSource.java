package apiserver.request;

import java.util.ArrayList;

public class SetAsyncRequestSequenceRequestWithSource {
    private String clusterName;

    private String sourceName;

    private ArrayList<String> svcList;

    public SetAsyncRequestSequenceRequestWithSource() {
        //do nothing
    }

    public SetAsyncRequestSequenceRequestWithSource(String clusterName, String sourceName, ArrayList<String> svcList) {
        this.clusterName = clusterName;
        this.sourceName = sourceName;
        this.svcList = svcList;
    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public String getSourceName() {
        return sourceName;
    }

    public void setSourceName(String sourceName) {
        this.sourceName = sourceName;
    }

    public ArrayList<String> getSvcList() {
        return svcList;
    }

    public void setSvcList(ArrayList<String> svcList) {
        this.svcList = svcList;
    }
}
