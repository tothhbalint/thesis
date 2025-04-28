const { Contract } = require("fabric-contract-api");
const crypto = require("crypto");

class KVContract extends Contract {
	constructor() {
		super("KVContract");
	}

	async instantiate() {
		// function that will be invoked on chaincode instantiation
	}

	async put(ctx, key, value) {
		await ctx.stub.putState(key, Buffer.from(value));
		return { success: "OK" };
	}

	async get(ctx, key) {
		const buffer = await ctx.stub.getState(key);
		if (!buffer || !buffer.length) return { error: "NOT_FOUND" };
		return { success: buffer.toString() };
	}

	async delete(ctx, key) {
		const exists = await ctx.stub.getState(key);
		if (!exists || !exists.length) {
			return { error: "NOT_FOUND" };
		}
		await ctx.stub.deleteState(key);
		return { success: "DELETED" };
	}

	async putPrivateMessage(ctx, collection) {
		const transient = ctx.stub.getTransient();
		const message = transient.get("message");
		await ctx.stub.putPrivateData(collection, "message", message);
		return { success: "OK" };
	}

	async getPrivateMessage(ctx, collection) {
		const message = await ctx.stub.getPrivateData(collection, "message");
		const messageString = message.toBuffer ? message.toBuffer().toString() : message.toString();
		return { success: messageString };
	}

	async verifyPrivateMessage(ctx, collection) {
		const transient = ctx.stub.getTransient();
		const message = transient.get("message");
		const messageString = message.toBuffer ? message.toBuffer().toString() : message.toString();
		const currentHash = crypto.createHash("sha256").update(messageString).digest("hex");
		const privateDataHash = (await ctx.stub.getPrivateDataHash(collection, "message")).toString("hex");
		if (privateDataHash !== currentHash) {
			return { error: "VERIFICATION_FAILED" };
		}
		return { success: "OK" };
	}

	async getStateByRange(ctx, startKey, endKey) {
		const iterator = await ctx.stub.getStateByRange(startKey, endKey);
		const results = [];
		let res = { done: false };
		while (!res.done) {
			res = await iterator.next();
			if (res.value && res.value.value.toString()) {
				results.push({
					key: res.value.key,
					value: res.value.value.toString(),
				});
			}
		}
		await iterator.close();
		return { success: results };
	}

	async ttest(ctx) { }


	async test(ctx) {
		const stub = ctx.stub;

		// Simple Ledger Calls
		let res;

		res = await stub.getState('someKey');
		console.log('GetState:', res.toString());

		await stub.putState('someKey', Buffer.from('someValue'));
		console.log('PutState: success');

		await stub.deleteState('someKey');
		console.log('DelState: success');

		const rangeIterator = await stub.getStateByRange('', '');
		while (true) {
			const r = await rangeIterator.next();
			if (r.value) {
				console.log('GetStateByRange result:', r.value.key, r.value.value.toString('utf8'));
			}
			if (r.done) {
				await rangeIterator.close();
				break;
			}
		}

		const queryIterator = await stub.getQueryResult('{"selector":{"docType":"someType"}}');
		while (true) {
			const r = await queryIterator.next();
			if (r.value) {
				console.log('GetQueryResult result:', r.value.key, r.value.value.toString('utf8'));
			}
			if (r.done) {
				await queryIterator.close();
				break;
			}
		}

		const historyIterator = await stub.getHistoryForKey('someKey');
		while (true) {
			const r = await historyIterator.next();
			if (r.value) {
				console.log('GetHistoryForKey result:', r.value.txId, r.value.value.toString('utf8'));
			}
			if (r.done) {
				await historyIterator.close();
				break;
			}
		}

		// Composite Key usage
		const ckey = stub.createCompositeKey('objectType', ['attr1', 'attr2']);
		console.log('CreateCompositeKey:', ckey);

		const split = stub.splitCompositeKey(ckey);
		console.log('SplitCompositeKey:', split);

		// Partial Composite Key
		const partialIterator = await stub.getStateByPartialCompositeKey('objectType', ['attr1']);
		while (true) {
			const r = await partialIterator.next();
			if (r.value) {
				console.log('PartialCompositeKey result:', r.value.key, r.value.value.toString('utf8'));
			}
			if (r.done) {
				await partialIterator.close();
				break;
			}
		}

		// Partial Composite Key with Pagination
		const { iterator, metadata } = await stub.getStateByPartialCompositeKeyWithPagination('objectType', ['attr1'], 1, '');
		while (true) {
			const r = await iterator.next();
			if (r.value) {
				console.log('Paginated PartialCompositeKey result:', r.value.key, r.value.value.toString('utf8'));
			}
			if (r.done) {
				await iterator.close();
				break;
			}
		}
		console.log('Pagination metadata:', metadata);
	}

}

exports.contracts = [KVContract];
