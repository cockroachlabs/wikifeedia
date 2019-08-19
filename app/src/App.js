import React from 'react';
// import logo from './logo.svg';
import './App.css';

import { ApolloClient } from 'apollo-client';
import { HttpLink } from 'apollo-link-http';
import { InMemoryCache } from 'apollo-cache-inmemory';
import gql from 'graphql-tag';
import { useQuery, ApolloProvider as ApolloHooksProvider } from '@apollo/react-hooks';
import { ApolloProvider } from 'react-apollo';

const client = new ApolloClient({
  // By default, this client will send queries to the
  //  `/graphql` endpoint on the same host
  // Pass the configuration option { uri: YOUR_GRAPHQL_API_URL } to the `HttpLink` to connect
  // to a different host
  link: new HttpLink({ uri: "/graphqlhttp" }),
  cache: new InMemoryCache(),
});

const GET_FEED = gql`
query Feed($project: String!) {
  articles(project: $project) {
    project
    abstract
    article
    articleURL
    dailyViews
    imageURL
    thumbnailURL
    title
  }

}`;

function Feed({ client, project }) {
  const {loading, error, data } = useQuery(GET_FEED, {
    variables: {
      "project": project,
    }
  });
  if (loading) return 'Loading...';
  if (error) return `Error! ${error.message}`;
  const rows = data.articles.map(({
    project,
    article,
    abstract,
    articleURL,
    dailyViews,
    imageURL,
    thumbnailURL,
    title
  }) => (
    <div className="Article" id={article} key={title} project={project}>
      <div className="ArticleImageContainer">
        <div className="ArticleImage">
          <a href={articleURL} target="_blank" rel="noopener noreferrer">
            <img src={thumbnailURL} alt={title}/>
          </a>
        </div>
      </div>
      <div className="ArticleContent">
        <h2 className="ArticleTitle">{title}</h2>
        <p className="ArticleAbstractText">
          {abstract}
        </p>
      </div>
    </div>
  ))
  return (
    <div className="Articles">
      {rows}
    </div>
  );
}

const languages = [
  "en",
	"fr",
	"es",
	"de",
	"ru",
	"ja",
	"nl",
	"it",
	"sv",
	"pl",
	"vi",
	"pt",
	"ar",
	"zh",
  "uk",
  "ro",
  "bg",
  "th"
]



class App extends React.Component {
  state = {
    project: "es",
  }
  setProject(project) {
    this.setState({project: project});
  }
  render()  {
    const langTabs = languages.map((lang) => {
      const onClick = () => {this.setProject(lang)};
      return (<div onClick={onClick}>{lang}</div>)
    });
    function App({project}) {
      return (
      <div className="App">
        <div className="LangTabs">
          {langTabs}
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

