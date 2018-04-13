package deltabackend.domain.sequenceDelta;

import java.util.ArrayList;

public class SetAsyncRequestSequenceRequestWithSource {

    private String sourceName;

    private ArrayList<String> svcList;

    public SetAsyncRequestSequenceRequestWithSource() {
        //do nothing
    }

    public SetAsyncRequestSequenceRequestWithSource(String sourceName, ArrayList<String> svcList) {
        this.sourceName = sourceName;
        this.svcList = svcList;
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
