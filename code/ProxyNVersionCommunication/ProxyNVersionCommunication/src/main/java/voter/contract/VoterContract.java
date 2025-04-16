package voter.contract;

import hu.bme.mit.ftsrg.hypernate.middleware.MiddlewareInfo;
import hu.bme.mit.ftsrg.hypernate.middleware.VoterMiddleware;
import org.hyperledger.fabric.contract.Context;
import org.hyperledger.fabric.contract.annotation.Contract;
import org.hyperledger.fabric.contract.annotation.Default;
import org.hyperledger.fabric.contract.annotation.Transaction;
import org.hyperledger.fabric.shim.ChaincodeException;
import org.hyperledger.fabric.shim.ChaincodeStub;
import voter.helpers.NVersion;
import voter.specification.IVotingContract;

import java.util.HashMap;


@Contract(name = "VoterContract")
@MiddlewareInfo(VoterMiddleware.class)
@Default
public class VoterContract implements IVotingContract {

    public HashMap<String,IVotingContract> contracts = new HashMap<String,IVotingContract>();

    public VoterContract(){
        contracts.put("Version1",new n_versions.version_1.SimpleContract());
        contracts.put("Version2",new n_versions.version_2.SimpleContract());
        contracts.put("Version3",new n_versions.version_3.SimpleContract());
    }

    @Transaction
    @NVersion
    public void createAsset(Context ctx, String assetId, String value){}

    @Transaction
    public String readAsset(Context ctx, String assetId){
        ChaincodeStub stub = ctx.getStub();
        String assetState = stub.getStringState(assetId);
        if (assetState.isEmpty()) {
            throw new ChaincodeException("Asset " + assetId + " does not exist");
        }
        return assetState;
    }

    @Transaction
    @NVersion
    public void updateAsset(Context ctx, String assetId, String newValue) {}

    @Transaction
    @NVersion
    public void deleteAsset(Context ctx, String assetId) {}

    @Transaction
    @NVersion
    public void queryAsset(Context ctx) {}

    @Transaction
    public void setupQueryAsset(Context ctx){
        for(int i = 1; i < 6; i++) {
            String key = "test_" + i;
            ctx.getStub().putState(key,"randval".getBytes(  ));
        }
    }
}
