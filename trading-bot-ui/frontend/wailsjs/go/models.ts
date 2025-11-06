export namespace database {
	
	export class Position {
	    id: number;
	    symbol: string;
	    quantity: number;
	    entry_price: number;
	    // Go type: time
	    entry_time: any;
	    exit_price?: number;
	    // Go type: time
	    exit_time?: any;
	    strategy: string;
	    is_open: boolean;
	    profit_loss?: number;
	    profit_loss_percent?: number;
	    buy_trade_id: number;
	    sell_trade_id?: number;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.symbol = source["symbol"];
	        this.quantity = source["quantity"];
	        this.entry_price = source["entry_price"];
	        this.entry_time = this.convertValues(source["entry_time"], null);
	        this.exit_price = source["exit_price"];
	        this.exit_time = this.convertValues(source["exit_time"], null);
	        this.strategy = source["strategy"];
	        this.is_open = source["is_open"];
	        this.profit_loss = source["profit_loss"];
	        this.profit_loss_percent = source["profit_loss_percent"];
	        this.buy_trade_id = source["buy_trade_id"];
	        this.sell_trade_id = source["sell_trade_id"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Trade {
	    id: number;
	    symbol: string;
	    side: string;
	    quantity: number;
	    price: number;
	    total: number;
	    strategy: string;
	    indicator_values: string;
	    signal_reason: string;
	    paper_trade: boolean;
	    // Go type: time
	    timestamp: any;
	    binance_order_id?: string;
	    profit_loss?: number;
	    profit_loss_percent?: number;
	    related_buy_id?: number;
	
	    static createFrom(source: any = {}) {
	        return new Trade(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.symbol = source["symbol"];
	        this.side = source["side"];
	        this.quantity = source["quantity"];
	        this.price = source["price"];
	        this.total = source["total"];
	        this.strategy = source["strategy"];
	        this.indicator_values = source["indicator_values"];
	        this.signal_reason = source["signal_reason"];
	        this.paper_trade = source["paper_trade"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.binance_order_id = source["binance_order_id"];
	        this.profit_loss = source["profit_loss"];
	        this.profit_loss_percent = source["profit_loss_percent"];
	        this.related_buy_id = source["related_buy_id"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TradeSummary {
	    total_trades: number;
	    total_buys: number;
	    total_sells: number;
	    total_profit_loss: number;
	    win_rate: number;
	    average_profit_loss: number;
	    largest_win: number;
	    largest_loss: number;
	    // Go type: time
	    start_date: any;
	    // Go type: time
	    end_date: any;
	
	    static createFrom(source: any = {}) {
	        return new TradeSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_trades = source["total_trades"];
	        this.total_buys = source["total_buys"];
	        this.total_sells = source["total_sells"];
	        this.total_profit_loss = source["total_profit_loss"];
	        this.win_rate = source["win_rate"];
	        this.average_profit_loss = source["average_profit_loss"];
	        this.largest_win = source["largest_win"];
	        this.largest_loss = source["largest_loss"];
	        this.start_date = this.convertValues(source["start_date"], null);
	        this.end_date = this.convertValues(source["end_date"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class BotStatus {
	    running: boolean;
	    strategy: string;
	    symbol: string;
	    trading_mode: string;
	    position?: database.Position;
	    last_trade?: database.Trade;
	
	    static createFrom(source: any = {}) {
	        return new BotStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.strategy = source["strategy"];
	        this.symbol = source["symbol"];
	        this.trading_mode = source["trading_mode"];
	        this.position = this.convertValues(source["position"], database.Position);
	        this.last_trade = this.convertValues(source["last_trade"], database.Trade);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CandleData {
	    timestamp: number;
	    open: number;
	    high: number;
	    low: number;
	    close: number;
	    volume: number;
	
	    static createFrom(source: any = {}) {
	        return new CandleData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.open = source["open"];
	        this.high = source["high"];
	        this.low = source["low"];
	        this.close = source["close"];
	        this.volume = source["volume"];
	    }
	}
	export class IndicatorData {
	    timestamp: number;
	    rsi: number;
	    macd: number;
	    signal: number;
	    histogram: number;
	    bb_upper: number;
	    bb_middle: number;
	    bb_lower: number;
	
	    static createFrom(source: any = {}) {
	        return new IndicatorData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.rsi = source["rsi"];
	        this.macd = source["macd"];
	        this.signal = source["signal"];
	        this.histogram = source["histogram"];
	        this.bb_upper = source["bb_upper"];
	        this.bb_middle = source["bb_middle"];
	        this.bb_lower = source["bb_lower"];
	    }
	}
	export class StrategyInfo {
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new StrategyInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class TimeframeChartData {
	    timeframe: string;
	    candles: CandleData[];
	    indicators: IndicatorData;
	    is_ready: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TimeframeChartData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timeframe = source["timeframe"];
	        this.candles = this.convertValues(source["candles"], CandleData);
	        this.indicators = this.convertValues(source["indicators"], IndicatorData);
	        this.is_ready = source["is_ready"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WalletBalance {
	    asset: string;
	    free: string;
	    locked: string;
	    usd_value: number;
	
	    static createFrom(source: any = {}) {
	        return new WalletBalance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.asset = source["asset"];
	        this.free = source["free"];
	        this.locked = source["locked"];
	        this.usd_value = source["usd_value"];
	    }
	}

}

