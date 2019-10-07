import React from 'react';
import './App.css';

import { Feed } from './Feed.js'

import { ApolloClient } from 'apollo-client';
import { HttpLink } from 'apollo-link-http';
import { InMemoryCache } from 'apollo-cache-inmemory';
import { ApolloProvider as ApolloHooksProvider } from '@apollo/react-hooks';
import { ApolloProvider } from 'react-apollo';

const client = new ApolloClient({
  // By default, this client will send queries to the
  //  `/graphql` endpoint on the same host
  // Pass the configuration option { uri: YOUR_GRAPHQL_API_URL } to the `HttpLink` to connect
  // to a different host
  link: new HttpLink({ uri: "/graphqlhttp" }),
  cache: new InMemoryCache(),
});

const languages = [
  "en", "fr", "es", "de", "ru", "ja", "nl", "it", "sv", "pl", "vi", "pt", "ar",
  "zh", "uk", "ro", "bg", "th"
];

function langTabs(project, setProject) {
  return languages.map((lang) => {
    const onClick = () => {setProject(lang)};
    let className = 'LangTab';
    if (lang === project) {
      className += ' LangTab-selected';
    }
    return (
      <div className={className} onClick={onClick} key={lang}>
        {lang}
      </div>
    );
  });
}

class App extends React.Component {
  state = {
    project: "en",
  }
  setProject(project) {
    this.setState({project: project});
  }
  render()  {
    const tabs = langTabs(this.state.project, (project) => { 
        this.setProject(project);
    });
    function App({project}) {
      return (
      <div className="App">
        <div className="LangTabs">
          {tabs}
        </div>
        <div className="AppHeaderContainer">
          <div className="App-header">
            <h1 className="AppTitle">Wikifeedia</h1> 
          </div>
        </div>
        
        <Feed project={project}/>
      </div>
      );
    };
    return (
      <ApolloProvider client={client}>
        <ApolloHooksProvider client={client}>
          <App project={this.state.project}/>
        </ApolloHooksProvider>
      </ApolloProvider>
    );
  }
}

export default App;

