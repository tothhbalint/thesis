package voter.helpers;

import hu.bme.mit.ftsrg.hypernate.context.HypernateContext;
import hu.bme.mit.ftsrg.hypernate.middleware.notification.VotingBegin;
import hu.bme.mit.ftsrg.hypernate.middleware.notification.VotingEnd;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Pointcut;
import voter.specification.IVotingContract;

import java.lang.reflect.Method;
import java.lang.reflect.Parameter;
import java.util.HashMap;

@Aspect
public class NVersionAspect {

    @Pointcut("@annotation(voter.helpers.NVersion)")
    public void nVersionPointcut() {
    }

    @Around("nVersionPointcut()")
    public Object aroundNVersion(ProceedingJoinPoint joinPoint) throws Throwable {
        Object[] args = joinPoint.getArgs();
        HypernateContext ctx = (HypernateContext) args[0];
        String methodName = joinPoint.getSignature().getName();

        HashMap<String, IVotingContract> contracts = (HashMap<String, IVotingContract>) joinPoint.getTarget().getClass().getField("contracts").get(joinPoint.getTarget());

        for (String version : contracts.keySet()) {
            ctx.notify(new VotingBegin(version));
            try{
                Method[] methods = contracts.get(version).getClass().getMethods();
                Method targetMethod = getMethod(methods, methodName, args);
                targetMethod.invoke(contracts.get(version), args);
            } catch (NoSuchMethodException e) {
                throw new RuntimeException("Method not found");
            } catch (Exception e) {
                throw new RuntimeException("Error invoking contract");
            }
        }
        ctx.notify(new VotingEnd());
        return joinPoint.proceed();
    }

    private static Method getMethod(Method[] methods, String methodName, Object[] args) throws NoSuchMethodException {
        Method targetMethod = null;

        for (Method method : methods){
            if (method.getName().equals(methodName) && method.getParameterTypes().length == args.length){
                boolean match = true;
                Parameter[] params = method.getParameters();
                for (int i = 0; i < params.length; i++){
                    if(!params[i].getType().isInstance(args[i])){
                        match = false;
                        break;
                    }
                }
                if (match) {
                    targetMethod = method;
                    break;
                }
            }
        }

        if (targetMethod == null){
            throw new NoSuchMethodException(methodName);
        }

        return targetMethod;
    }
}
