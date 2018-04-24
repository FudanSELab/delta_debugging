package deltabackend.domain.configDelta;

import java.util.ArrayList;
import java.util.List;

public class CMConfig {
    private String type;
    private List<CM> values = new ArrayList<>();

    public CMConfig(){

    }

    public void addValues(CM c){
        values.add(c);
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public List<CM> getValues() {
        return values;
    }

    public void setValues(List<CM> values) {
        this.values = values;
    }

    public String toString(){
        return this.type + ": " + this.values;
    }
}
