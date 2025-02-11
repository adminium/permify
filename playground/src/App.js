import React from "react";
import AppRouter from "./router/Router";
import {Layout} from "./utility/context/Layout"
import {Provider} from 'react-redux'
import {PersistGate} from "redux-persist/integration/react";

function App({store}) {
    return (
        <Provider store={store}>
                <Layout>
                    <AppRouter/>
                </Layout>
        </Provider>
    );
}

export default App
