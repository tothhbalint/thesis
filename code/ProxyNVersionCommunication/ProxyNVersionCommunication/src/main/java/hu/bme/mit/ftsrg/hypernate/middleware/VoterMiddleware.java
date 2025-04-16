package hu.bme.mit.ftsrg.hypernate.middleware;

import hu.bme.mit.ftsrg.hypernate.middleware.notification.HypernateNotification;
import hu.bme.mit.ftsrg.hypernate.middleware.notification.VotingBegin;
import hu.bme.mit.ftsrg.hypernate.middleware.notification.VotingEnd;
import org.hyperledger.fabric.shim.ChaincodeException;
import org.hyperledger.fabric.shim.ChaincodeStub;
import org.hyperledger.fabric.shim.ledger.KeyValue;
import org.hyperledger.fabric.shim.ledger.QueryResultsIterator;
import org.slf4j.Logger;

import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.function.Function;

public class VoterMiddleware extends StubMiddleware {
    private String _invoker_name = "";
    private boolean _voting = false;

    private final Logger logger = org.slf4j.LoggerFactory.getLogger(VoterMiddleware.class);

    static class VersionData {
        private final Set<String> readSet;
        private final Map<String, String> writeSet;

        public VersionData(Set<String> readSet, Map<String, String> writeSet) {
            this.readSet = new HashSet<>(readSet);
            this.writeSet = new HashMap<>(writeSet);
        }

        public VersionData() {
            this.readSet = new HashSet<>();
            this.writeSet = new HashMap<>();
        }

        public void addRead(String key) {
            readSet.add(key);
        }

        public void addWrite(String key, String value) {
            writeSet.put(key, value);
        }

        public void addWrite(String key, byte[] value) {
            writeSet.put(key, new String(value));
        }

        public Set<String> getReadSet() {
            return readSet;
        }

        public Map<String, String> getWriteSet() {
            return writeSet;
        }

        public boolean compareRWSet(VersionData other) {
            for (String key : readSet) {
                if (!other.readSet.contains(key)) {
                    return false;
                }
            }
            for (String key : writeSet.keySet()) {
                if (other.writeSet.containsKey(key)) {
                    if (!other.writeSet.get(key).equals(writeSet.get(key))) {
                        return false;
                    }
                } else {
                    return false;
                }
            }
            return true;
        }
    }

    private final HashMap<String, VersionData> rwSets = new HashMap<>();

    @Override
    public void putState(String key, final byte[] value) {
        if (!_voting) {
            this.nextStub.putState(key, value);
        } else {
            rwSets.get(_invoker_name).addWrite(key, value);
        }
    }

    //caching for getState
    @Override
    public byte[] getState(String key) {
        if (!_voting) {
            return this.nextStub.getState(key);
        } else {
            byte[] read_value = this.nextStub.getState(key);
            rwSets.get(_invoker_name).addRead(key);
            return read_value;
        }
    }

    @Override
    public void putStringState(String key, final String value) {
        if (!_voting) {
            this.nextStub.putStringState(key, value);
        } else {
            rwSets.get(_invoker_name).addWrite(key, value);
        }
    }

    @Override
    public String getStringState(String key) {
        if (!_voting) {
            return this.nextStub.getStringState(key);
        } else {
            String read_value = this.nextStub.getStringState(key);
            rwSets.get(_invoker_name).addRead(key);
            return read_value;
        }
    }

    public QueryResultsIterator<KeyValue> queryFuncBody(QueryResultsIterator<KeyValue> iterator) {
        if (!_voting) {
            return iterator;
        } else{
            iterator.forEach(kv -> rwSets.get(_invoker_name).addRead(kv.getKey()));
            return iterator;
        }
    }

    @Override
    public QueryResultsIterator<KeyValue> getStateByRange(String var1, String var2) {
        return queryFuncBody(this.nextStub.getStateByRange(var1,var2));
    }

    @Override
    public QueryResultsIterator<KeyValue> getStateByPartialCompositeKey(String key) {
        return queryFuncBody(this.nextStub.getStateByPartialCompositeKey(key));
    }


    @Override
    public void delState(String key) {
        this.nextStub.delState(key);
    }

    @Override
    public void onNotification(HypernateNotification notification) {
        if (notification instanceof VotingBegin) {
            _invoker_name = ((VotingBegin) notification).getVoterName();
            rwSets.put(_invoker_name, new VersionData());
            _voting = true;
        } else if (notification instanceof VotingEnd) {
            _voting = false;
            onVotingEnd();
        }
    }

    public void onVotingEnd() {
        int maxVoteCount = 0;
        VersionData majorityReadWriteSet = new VersionData();
        HashMap<VersionData, Integer> voteCounts = new HashMap<>();

        for (VersionData rwSet : rwSets.values()) {
            voteCounts.put(rwSet, voteCounts.getOrDefault(rwSet, 0) + 1);
            if (voteCounts.get(rwSet) > maxVoteCount) {
                maxVoteCount = voteCounts.get(rwSet);
                majorityReadWriteSet = rwSet;
            }
        }

        List<String> losingVersions = new ArrayList<>();
        VersionData finalMajorityReadWriteSet = majorityReadWriteSet;
        rwSets.forEach((key, value) -> {
            if (!finalMajorityReadWriteSet.compareRWSet(value)) {
                losingVersions.add(key);
            }
        });

        if (!losingVersions.isEmpty()) {
            if (losingVersions.size() > rwSets.size() / 2) {
                logger.error("MULTIPLE INCORRECT VERSIONS: {}", losingVersions);
                throw new ChaincodeException("N-Version ERROR", "Multiple incorrect versions can't commit to ledger");
            } else {
                logger.warn("INCORRECT VERSIONS: {}", losingVersions);
            }
        }
        majorityReadWriteSet.writeSet.forEach((key, value) -> {
            putState(key, value.getBytes());
        });
    }
}