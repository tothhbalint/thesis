package hu.bme.mit.ftsrg.hypernate.middleware.notification;

public class VotingBegin extends HypernateNotification{
    private final String voter_name;

    public VotingBegin(String voter_name){
        this.voter_name = voter_name;
    }

    public String getVoterName(){
        return voter_name;
    }
}
