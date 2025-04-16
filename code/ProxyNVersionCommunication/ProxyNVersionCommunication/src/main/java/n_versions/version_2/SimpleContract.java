package n_versions.version_2;

import org.hyperledger.fabric.contract.Context;
import org.hyperledger.fabric.contract.ContractInterface;
import org.hyperledger.fabric.contract.annotation.Contract;
import org.hyperledger.fabric.shim.ChaincodeStub;
import voter.specification.IVotingContract;

import java.util.Arrays;


@Contract(name = "SimpleContract")
public class SimpleContract implements IVotingContract {
    public SimpleContract() {
    }

    public void createAsset(Context ctx, String assetId, String value) {
        ChaincodeStub stub = ctx.getStub();
        byte[] assetState = stub.getState(assetId);
        if (assetState != null && assetState.length > 0) {
            throw new RuntimeException("Asset " + assetId + " already exists");
        }
        stub.putState(assetId, value.getBytes());
    }

    public void updateAsset(Context ctx, String assetId, String newValue) {
        ChaincodeStub stub = ctx.getStub();
        String assetState = Arrays.toString(stub.getState(assetId));
        if (assetState.isEmpty()) {
            throw new RuntimeException("Asset " + assetId + " does not exist");
        }
        stub.putState(assetId, newValue.getBytes());
    }

    public void deleteAsset(Context ctx, String assetId) {
        ChaincodeStub stub = ctx.getStub();
        String assetState = Arrays.toString(stub.getState(assetId));
        if (assetState.isEmpty()) {
            throw new RuntimeException("Asset " + assetId + " does not exist");
        }
        stub.delState(assetId);
    }

    public void queryAsset(Context ctx) {
        ChaincodeStub stub = ctx.getStub();
        stub.getStateByRange("test_1","test_3");
    }
}
