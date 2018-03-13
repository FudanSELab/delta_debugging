package backend.service;

import net.sf.json.JSONArray;
import net.sf.json.JSONObject;
import net.sf.json.JsonConfig;
import org.springframework.stereotype.Service;

import java.io.*;
import java.util.*;

@Service
public class ConfigService {

//    private String jsonPath = this.getClass().getResource("/").getPath() + "testConfig.json";
//    private String jsonPath = "/testConfig.json";
    private String jsonPath = "/docker-testConfig.json";
    private String fileString;
    private String testDir;
    private String classDir;
    private JSONObject testCases;
    Map<String, List<String>> testMap = new HashMap<String, List<String>>();

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
        if(null != testCases){
            testMap.clear();
            Iterator<String> sIterator = testCases.keys();
            while(sIterator.hasNext()){
                // 获得key
                String key = sIterator.next();
                // 根据key获得value, value也可以是JSONObject,JSONArray,使用对应的参数接收即可
                JSONArray value = testCases.getJSONArray(key);
                testMap.put(key, JSONArray.toList(value, String.class, new JsonConfig()));
            }
        }
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

    public void reloadJson(){
        fileString = null;
        testDir = null;
        classDir = null;
        testCases = null;
        init();
    }

    public boolean containTestCase(String s){
        if( null == testCases){
            init();
        }
        for(String key: testMap.keySet()){
            List<String> names = testMap.get(key);
            for(String name : names) {
                if (name.equals(s)) {
                    return true;
                }
            }
        }
        return false;
    }

    public Map<String, List<String>> getTestFileNames(){
//        Map<String, List<String>> testMap = new HashMap<String, List<String>>();
        if(testCases == null){
           init();
        }
//        Iterator<String> sIterator = testCases.keys();
//        while(sIterator.hasNext()){
//            // 获得key
//            String key = sIterator.next();
//            // 根据key获得value, value也可以是JSONObject,JSONArray,使用对应的参数接收即可
//            JSONArray value = testCases.getJSONArray(key);
//            testMap.put(key, JSONArray.toList(value, String.class, new JsonConfig()));
//        }
        return testMap;
    }

}
