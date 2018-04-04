package apiserver.request;

import java.util.ArrayList;

public class SetAsyncRequestSequenceRequest {

    private ArrayList<String> svcList;

    public SetAsyncRequestSequenceRequest() {
        //do nothing
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
