package apiserver.util;

public class LongRuleConfig {

    public String svcName;

    public LongRuleConfig() {
        //do nothing
    }

    public LongRuleConfig(String svcName) {
        this.svcName = svcName;
    }

    public String getSvcName() {
        return svcName;
    }

    public void setSvcName(String svcName) {
        this.svcName = svcName;
    }
}
