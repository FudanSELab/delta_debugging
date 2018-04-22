package deltabackend.domain.sequenceDelta;

import java.util.ArrayList;
import java.util.List;

public class SingleSequenceDelta {

    String sender;
    ArrayList<String> receivers;

    public SingleSequenceDelta(){

    }


    public String getSender() {
        return sender;
    }

    public void setSender(String sender) {
        this.sender = sender;
    }

    public ArrayList<String> getReceivers() {
        return receivers;
    }

    public void setReceivers(ArrayList<String> receivers) {
        this.receivers = receivers;
    }


}
