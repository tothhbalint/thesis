package hu.bme.mit.ftsrg.hypernate.middleware;

import static org.junit.jupiter.api.Assertions.*;

import hu.bme.mit.ftsrg.hypernate.middleware.VoterMiddleware.VersionData;
import hu.bme.mit.ftsrg.hypernate.middleware.notification.VotingBegin;
import hu.bme.mit.ftsrg.hypernate.middleware.notification.VotingEnd;
import org.hyperledger.fabric.shim.ChaincodeException;
import org.hyperledger.fabric.shim.ChaincodeStub;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.*;

@ExtendWith(MockitoExtension.class)
class VoterMiddlewareTest {
    @Mock
    ChaincodeStub fabricStub;

    StubMiddlewareChain.Builder builder;


    @BeforeEach
    void setUp() {
        builder = StubMiddlewareChain.builder(fabricStub);
        builder.push(VoterMiddleware.class);
    }

    void send_stub_requests(String name,VersionData versionData, VoterMiddleware stub) {
        stub.onNotification(new VotingBegin(name));
        versionData.getReadSet().forEach(stub::getState);
        versionData.getWriteSet().forEach((key,value) -> stub.putState(key,value.getBytes()));
    }

    @Test
    void check_voting_when_one_difference() {
        // 1 different results -> Expected: pass

        final StubMiddlewareChain chain = builder.build();
        final VoterMiddleware voterMiddleware = (VoterMiddleware) chain.getFirst();
        final HashMap<String,VersionData> versions = new HashMap<>();

        // insert three versions write calls
        versions.put("v1", new VersionData(Set.of(), Map.of("test_key", "test_value")));
        versions.put("v2", new VersionData(Set.of(), Map.of("test_key", "test_value")));
        versions.put("v3", new VersionData(Set.of(), Map.of("test_key", "wrong_value")));

        versions.forEach((key,value) -> send_stub_requests(key, value, voterMiddleware));

        verify(fabricStub,never()).putState("test_key", "test_value".getBytes());
        voterMiddleware.onNotification(new VotingEnd());
        //check if the correct output has been voted
        verify(fabricStub).putState("test_key", "test_value".getBytes());
        // check that the wrong value hasn't been written
        verify(fabricStub,never()).putState("test_key", "wrong_value".getBytes());
    }

    @Test
    void check_voting_when_two_difference() {
        final StubMiddlewareChain chain = builder.build();
        final VoterMiddleware voterMiddleware = (VoterMiddleware) chain.getFirst();
        final HashMap<String,VersionData> versions = new HashMap<>();

        // insert three versions write calls
        // insert three versions write calls
        versions.put("v1", new VersionData(Set.of(), Map.of("test_key", "test_value")));
        versions.put("v2", new VersionData(Set.of(), Map.of("test_key", "wrong_value_1")));
        versions.put("v3", new VersionData(Set.of(), Map.of("test_key", "wrong_value_2")));

        versions.forEach((key,value) -> send_stub_requests(key, value, voterMiddleware));
        // check that no data has been written to the ledger
        verify(fabricStub,never()).putState("test_key", "test_value".getBytes());
        verify(fabricStub,never()).putState("test_key", "wrong_value_1".getBytes());
        verify(fabricStub,never()).putState("test_key", "wrong_value_2".getBytes());
    }
    @Test
    void check_read_set_diff_when_write_set_match() {
        // check read set differs, when the write sets also differ at that two

        final StubMiddlewareChain chain = builder.build();
        final VoterMiddleware voterMiddleware = (VoterMiddleware) chain.getFirst();
        final HashMap<String,VersionData> versions = new HashMap<>();

        // insert three versions write calls
        // insert three versions write calls
        versions.put("v1", new VersionData(Set.of("right_key"), Map.of("test_key", "test_value")));
        versions.put("v2", new VersionData(Set.of("right_key"), Map.of("test_key", "test_value")));
        versions.put("v3", new VersionData(Set.of("wrong_key"), Map.of("test_key", "wrong_value")));

        versions.forEach((key,value) -> send_stub_requests(key, value, voterMiddleware));
        voterMiddleware.onNotification(new VotingEnd());
        // check that no data has been written to the ledger
        verify(fabricStub).putState("test_key", "test_value".getBytes());
        verify(fabricStub,never()).putState("test_key", "wrong_value".getBytes());
    }

    @Test
    void check_read_set_diff_when_write_set_diff() {
        // test in the case that the read set differs at different versions, than where the write sets diff
        final StubMiddlewareChain chain = builder.build();
        final VoterMiddleware voterMiddleware = (VoterMiddleware) chain.getFirst();
        final HashMap<String,VersionData> versions = new HashMap<>();

        versions.put("v1", new VersionData(Set.of("right_key"), Map.of("test_key", "test_value")));
        versions.put("v2", new VersionData(Set.of("wrong_key"), Map.of("test_key", "test_value")));
        versions.put("v3", new VersionData(Set.of("right_key"), Map.of("test_key", "wrong_value")));

        versions.forEach((key,value) -> send_stub_requests(key, value, voterMiddleware));
        ChaincodeException exception = assertThrows(ChaincodeException.class, () -> {
            voterMiddleware.onNotification(new VotingEnd());
        });
        // check that no data has been written to the ledger
        verify(fabricStub, never()).putState("test_key", "test_value".getBytes());
        verify(fabricStub, never()).putState("test_key", "wrong_value".getBytes());
    }
}