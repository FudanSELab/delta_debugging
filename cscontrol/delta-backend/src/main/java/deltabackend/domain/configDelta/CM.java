package deltabackend.domain.configDelta;

public class CM {
    private String key;
    private String value;

    public CM(){

    }

    public CM(String key, String value){
        this.key = key;
        this.value = value;
    }

    public String getKey() {
        return key;
    }

    public void setKey(String key) {
        this.key = key;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public String toString(){
        return this.key + ": " + this.value;
    }
}
