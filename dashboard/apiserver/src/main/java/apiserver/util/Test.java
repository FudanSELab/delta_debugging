package apiserver.util;

import apiserver.bean.QueryDeploymentsListResponse;
import com.alibaba.fastjson.JSON;

import java.io.*;

public class Test {
    public static void main(String[] args){
        String encoding = "UTF-8";
        File file = new File("result.json");
        Long filelength = file.length();
        byte[] filecontent = new byte[filelength.intValue()];
        String json = "";
        try {
            FileInputStream in = new FileInputStream(file);
            in.read(filecontent);
            in.close();
        } catch (FileNotFoundException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
        try {
            json =  new String(filecontent, encoding);
        } catch (UnsupportedEncodingException e) {
            System.err.println("The OS does not support " + encoding);
            e.printStackTrace();
        }
        QueryDeploymentsListResponse deploymentsList = JSON.parseObject(json,QueryDeploymentsListResponse.class);
        System.out.println(deploymentsList.getItems().get(0).getMetadata().getName());
    }

}
