package backend.service;

import net.sf.json.JSONObject;
import org.springframework.stereotype.Service;

import java.io.*;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;

@Service
public class ConfigService {

//    private String jsonPath = this.getClass().getResource("/").getPath() + "testConfig.json";
//    private String jsonPath = "/testConfig.json";
    private String jsonPath = "/docker-testConfig.json";
    private String fileString;
    private String testDir;
    private String classDir;
    private JSONObject testCases;

    public void readFile(){
        BufferedReader reader = null;
        fileString = "";
        try{
//            FileInputStream fileInputStream = new FileInputStream();
//            InputStreamReader inputStreamReader = new InputStreamReader(fileInputStream, "UTF-8");
            InputStreamReader inputStreamReader = new InputStreamReader(this.getClass().getResourceAsStream(jsonPath));
            reader = new BufferedReader(inputStreamReader);
            String tempString = null;
            while((tempString = reader.readLine()) != null){
                fileString += tempString;
            }
            reader.close();
        }catch(IOException e){
            e.printStackTrace();
        }finally{
            if(reader != null){
                try {
                    reader.close();
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }
        }
    }

    public void convertJson(){
        JSONObject jsonObject = JSONObject.fromObject(fileString);
        testDir = jsonObject.getString("testDir");
        classDir = jsonObject.getString("classDir");
        testCases = jsonObject.getJSONObject("testCases");
    }

    public void init(){
        readFile();
        convertJson();
    }

    public String getFileString(){
        return this.fileString;
    }

    public String getTestDir(){
        if( null == testDir){
            init();
        }
        return testDir;
    }

    public String getClassDir(){
        if( null == classDir){
            init();
        }
        return classDir;
    }

    public boolean containTestCase(String s){
        if( null == testCases){
            init();
        }
        return testCases.containsKey(s);
    }

    public String getTestCase(String s){
        if( null == testCases){
            init();
        }
        if(testCases.containsKey(s)){
            return testCases.getString(s);
        }
        return null;
    }

    public void reloadJson(){
        fileString = null;
        testDir = null;
        classDir = null;
        testCases = null;
        init();
    }

    public List<String> getTestFileNames(){
        List<String> s = new ArrayList<String>();
        if(testCases == null){
           init();
        }
        Iterator<String> sIterator = testCases.keys();
        while(sIterator.hasNext()){
            // 获得key
            String key = sIterator.next();
            // 根据key获得value, value也可以是JSONObject,JSONArray,使用对应的参数接收即可
            String value = testCases.getString(key);
            s.add(key);
//            System.out.println("key: "+key+",value: "+value);
        }
        return s;
    }

}
