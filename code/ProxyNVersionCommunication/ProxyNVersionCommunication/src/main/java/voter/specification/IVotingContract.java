package voter.specification;

import hu.bme.mit.ftsrg.hypernate.contract.HypernateContract;
import org.hyperledger.fabric.contract.Context;

public interface IVotingContract extends HypernateContract {

    public void createAsset(Context ctx, String assetId, String value);

    public void updateAsset(Context ctx, String assetId, String newValue);

    public void deleteAsset(Context ctx, String assetId);

    public void queryAsset(Context ctx);
}
