package apiserver.bean;

import java.util.List;

public class CMConfig {
    private String type;
    private List<CM> values;

    public CMConfig(){

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
}
