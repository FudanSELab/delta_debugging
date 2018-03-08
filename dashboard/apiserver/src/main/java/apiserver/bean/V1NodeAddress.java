package apiserver.bean;

public class V1NodeAddress {
    private String address = null;
    private String type = null;

    public V1NodeAddress(){

    }

    public String getAddress() {
        return address;
    }

    public void setAddress(String address) {
        this.address = address;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }
}
